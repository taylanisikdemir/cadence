package election

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/fx"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/service/sharddistributor/config"
	"github.com/uber/cadence/service/sharddistributor/leader/process"
	"github.com/uber/cadence/service/sharddistributor/leader/store"
)

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination=election_mock.go Factory,Elector

var errSelfResign = fmt.Errorf("self-resigned")

// Module provides election factory for fx app.
var Module = fx.Module(
	"leader-election",
	fx.Provide(NewElectionFactory),
)

type ProcessFunc func(ctx context.Context) error

// Elector handles leader election for a specific namespace
type Elector interface {
	Run(ctx context.Context) <-chan bool
}

// Factory creates elector instances
type Factory interface {
	CreateElector(ctx context.Context, namespace string) (Elector, error)
}

type electionFactory struct {
	hostname       string
	cfg            config.Election
	store          store.Elector
	logger         log.Logger
	serviceID      string
	clock          clock.TimeSource
	processFactory process.Factory
}

type elector struct {
	hostname       string
	namespace      string
	store          store.Elector
	logger         log.Logger
	cfg            config.Election
	leaderStarted  time.Time
	clock          clock.TimeSource
	processFactory process.Factory
}

type FactoryParams struct {
	fx.In

	HostName       string `name:"hostname"`
	Cfg            config.LeaderElection
	Store          store.Elector
	Logger         log.Logger
	Clock          clock.TimeSource
	ProcessFactory process.Factory
}

// NewElectionFactory creates a new election factory
func NewElectionFactory(p FactoryParams) Factory {
	return &electionFactory{
		cfg:            p.Cfg.Election,
		store:          p.Store,
		logger:         p.Logger,
		clock:          p.Clock,
		hostname:       p.HostName,
		processFactory: p.ProcessFactory,
	}
}

// CreateElector creates a new elector for the given namespace
func (f *electionFactory) CreateElector(ctx context.Context, namespace string) (Elector, error) {
	return &elector{
		namespace:      namespace,
		store:          f.store,
		logger:         f.logger.WithTags(tag.ComponentLeaderElection, tag.ShardNamespace(namespace)),
		cfg:            f.cfg,
		clock:          f.clock,
		hostname:       f.hostname,
		processFactory: f.processFactory,
	}, nil
}

// Run starts the leader election process it returns a channel that will return the value if the current instance becomes the leader or resigns from leadership.
func (e *elector) Run(ctx context.Context) <-chan bool {
	leaderCh := make(chan bool, 1)

	// Create a child context that we can explicitly cancel when errors occur
	runCtx, cancelRun := context.WithCancel(ctx)

	go func() {
		defer close(leaderCh)
		defer cancelRun() // Ensure child context is canceled on exit
		defer func() {
			if r := recover(); r != nil {
				e.logger.Error("Panic in election process", tag.Value(r))
			}
		}()

		for {
			if err := e.runElection(runCtx, leaderCh); err != nil {
				// Check if parent context is already canceled
				if runCtx.Err() != nil {
					// Context is canceled, exit immediately without logging
					return
				}

				// Self resign, immediately retry, otherwise, wait
				if !errors.Is(err, errSelfResign) {
					e.logger.Error("Error in election, retrying", tag.Error(err))

					select {
					case <-runCtx.Done():
						return // Context was canceled, exit immediately
					case <-e.clock.After(e.cfg.FailedElectionCooldown):
						// Continue after cooldown
					}
				}
			}
			if runCtx.Err() != nil {
				break
			}
		}
	}()

	return leaderCh
}

// runElection runs a single election attempt
func (e *elector) runElection(ctx context.Context, leaderCh chan<- bool) (err error) {
	// Check if context is already canceled before proceeding
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled before election: %w", ctx.Err())
	}

	// Add random delay before campaigning to spread load across instances
	delay := time.Duration(rand.Intn(int(e.cfg.MaxRandomDelay)))

	e.logger.Debug("Adding random delay before campaigning", tag.ElectionDelay(delay))

	select {
	case <-e.clock.After(delay):
		// Continue after delay
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during pre-campaign delay: %w", ctx.Err())
	}

	leaderProcess := e.processFactory.CreateProcessor(e.namespace)

	election, err := e.store.CreateElection(ctx, e.namespace)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer func() {
		resignErr := e.resign(election, leaderProcess.Terminate)
		if resignErr != nil {
			if err == nil {
				err = resignErr
			} else {
				// Something already went wrong, so the process is most likely not started or stopped.
				// It should be safe to report leadership lost.
				e.logger.Error("Error resigning leader", tag.Error(resignErr))
			}
		}

		// We are no longer a leader and OnResign is called - notify the manager that leader elected process
		leaderCh <- false
	}()

	// Campaign to become leader
	if err := election.Campaign(ctx, e.hostname); err != nil {
		return fmt.Errorf("failed to campaign: %w", err)
	}

	err = leaderProcess.Run(ctx)
	if err != nil {
		return fmt.Errorf("onLeader: %w", err)
	}

	// Successfully became leader
	e.leaderStarted = e.clock.Now()
	leaderCh <- true

	e.logger.Info("Became leader")

	// Start a timer to voluntarily resign after the leadership period
	leaderTimer := e.clock.NewTimer(e.cfg.LeaderPeriod)
	defer leaderTimer.Stop()

	// Watch for session expiration, context cancellation, or timer expiration
	select {
	case <-ctx.Done():
		e.logger.Info("Context cancelled while leader")
		return nil

	case <-election.Done():
		e.logger.Info("Session expired while leader")
		return fmt.Errorf("session expired")

	case <-leaderTimer.Chan():
		e.logger.Info("Leadership period ended, voluntarily resigning")

		return errSelfResign
	}
}

func (e *elector) resign(election store.Election, onResign ProcessFunc) error {
	ctx, cancel := e.clock.ContextWithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// First try to call onResign
	resignErr := onResign(ctx)

	// Then try to resign leadership, regardless of whether onResign succeeded
	resignElectionErr := election.Resign(ctx)

	// Combine errors if both failed
	if resignErr != nil && resignElectionErr != nil {
		return fmt.Errorf("multiple errors: OnResign: %w; election resign: %v", resignErr, resignElectionErr)
	}

	// Return whichever error occurred
	if resignErr != nil {
		return fmt.Errorf("OnResign: %w", resignErr)
	}
	if resignElectionErr != nil {
		return fmt.Errorf("election resign: %w", resignElectionErr)
	}

	return nil
}
