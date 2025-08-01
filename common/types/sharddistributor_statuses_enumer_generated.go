// Code generated by "enumer -type=ExecutorStatus,ShardStatus,AssignmentStatus -json -output sharddistributor_statuses_enumer_generated.go"; DO NOT EDIT.

package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _ExecutorStatusName = "ExecutorStatusINVALIDExecutorStatusACTIVEExecutorStatusDRAININGExecutorStatusDRAINED"

var _ExecutorStatusIndex = [...]uint8{0, 21, 41, 63, 84}

const _ExecutorStatusLowerName = "executorstatusinvalidexecutorstatusactiveexecutorstatusdrainingexecutorstatusdrained"

func (i ExecutorStatus) String() string {
	if i < 0 || i >= ExecutorStatus(len(_ExecutorStatusIndex)-1) {
		return fmt.Sprintf("ExecutorStatus(%d)", i)
	}
	return _ExecutorStatusName[_ExecutorStatusIndex[i]:_ExecutorStatusIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _ExecutorStatusNoOp() {
	var x [1]struct{}
	_ = x[ExecutorStatusINVALID-(0)]
	_ = x[ExecutorStatusACTIVE-(1)]
	_ = x[ExecutorStatusDRAINING-(2)]
	_ = x[ExecutorStatusDRAINED-(3)]
}

var _ExecutorStatusValues = []ExecutorStatus{ExecutorStatusINVALID, ExecutorStatusACTIVE, ExecutorStatusDRAINING, ExecutorStatusDRAINED}

var _ExecutorStatusNameToValueMap = map[string]ExecutorStatus{
	_ExecutorStatusName[0:21]:       ExecutorStatusINVALID,
	_ExecutorStatusLowerName[0:21]:  ExecutorStatusINVALID,
	_ExecutorStatusName[21:41]:      ExecutorStatusACTIVE,
	_ExecutorStatusLowerName[21:41]: ExecutorStatusACTIVE,
	_ExecutorStatusName[41:63]:      ExecutorStatusDRAINING,
	_ExecutorStatusLowerName[41:63]: ExecutorStatusDRAINING,
	_ExecutorStatusName[63:84]:      ExecutorStatusDRAINED,
	_ExecutorStatusLowerName[63:84]: ExecutorStatusDRAINED,
}

var _ExecutorStatusNames = []string{
	_ExecutorStatusName[0:21],
	_ExecutorStatusName[21:41],
	_ExecutorStatusName[41:63],
	_ExecutorStatusName[63:84],
}

// ExecutorStatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ExecutorStatusString(s string) (ExecutorStatus, error) {
	if val, ok := _ExecutorStatusNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ExecutorStatusNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ExecutorStatus values", s)
}

// ExecutorStatusValues returns all values of the enum
func ExecutorStatusValues() []ExecutorStatus {
	return _ExecutorStatusValues
}

// ExecutorStatusStrings returns a slice of all String values of the enum
func ExecutorStatusStrings() []string {
	strs := make([]string, len(_ExecutorStatusNames))
	copy(strs, _ExecutorStatusNames)
	return strs
}

// IsAExecutorStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ExecutorStatus) IsAExecutorStatus() bool {
	for _, v := range _ExecutorStatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for ExecutorStatus
func (i ExecutorStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ExecutorStatus
func (i *ExecutorStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ExecutorStatus should be a string, got %s", data)
	}

	var err error
	*i, err = ExecutorStatusString(s)
	return err
}

const _ShardStatusName = "ShardStatusINVALIDShardStatusREADY"

var _ShardStatusIndex = [...]uint8{0, 18, 34}

const _ShardStatusLowerName = "shardstatusinvalidshardstatusready"

func (i ShardStatus) String() string {
	if i < 0 || i >= ShardStatus(len(_ShardStatusIndex)-1) {
		return fmt.Sprintf("ShardStatus(%d)", i)
	}
	return _ShardStatusName[_ShardStatusIndex[i]:_ShardStatusIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _ShardStatusNoOp() {
	var x [1]struct{}
	_ = x[ShardStatusINVALID-(0)]
	_ = x[ShardStatusREADY-(1)]
}

var _ShardStatusValues = []ShardStatus{ShardStatusINVALID, ShardStatusREADY}

var _ShardStatusNameToValueMap = map[string]ShardStatus{
	_ShardStatusName[0:18]:       ShardStatusINVALID,
	_ShardStatusLowerName[0:18]:  ShardStatusINVALID,
	_ShardStatusName[18:34]:      ShardStatusREADY,
	_ShardStatusLowerName[18:34]: ShardStatusREADY,
}

var _ShardStatusNames = []string{
	_ShardStatusName[0:18],
	_ShardStatusName[18:34],
}

// ShardStatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ShardStatusString(s string) (ShardStatus, error) {
	if val, ok := _ShardStatusNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ShardStatusNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ShardStatus values", s)
}

// ShardStatusValues returns all values of the enum
func ShardStatusValues() []ShardStatus {
	return _ShardStatusValues
}

// ShardStatusStrings returns a slice of all String values of the enum
func ShardStatusStrings() []string {
	strs := make([]string, len(_ShardStatusNames))
	copy(strs, _ShardStatusNames)
	return strs
}

// IsAShardStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ShardStatus) IsAShardStatus() bool {
	for _, v := range _ShardStatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for ShardStatus
func (i ShardStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ShardStatus
func (i *ShardStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ShardStatus should be a string, got %s", data)
	}

	var err error
	*i, err = ShardStatusString(s)
	return err
}

const _AssignmentStatusName = "AssignmentStatusINVALIDAssignmentStatusREADY"

var _AssignmentStatusIndex = [...]uint8{0, 23, 44}

const _AssignmentStatusLowerName = "assignmentstatusinvalidassignmentstatusready"

func (i AssignmentStatus) String() string {
	if i < 0 || i >= AssignmentStatus(len(_AssignmentStatusIndex)-1) {
		return fmt.Sprintf("AssignmentStatus(%d)", i)
	}
	return _AssignmentStatusName[_AssignmentStatusIndex[i]:_AssignmentStatusIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _AssignmentStatusNoOp() {
	var x [1]struct{}
	_ = x[AssignmentStatusINVALID-(0)]
	_ = x[AssignmentStatusREADY-(1)]
}

var _AssignmentStatusValues = []AssignmentStatus{AssignmentStatusINVALID, AssignmentStatusREADY}

var _AssignmentStatusNameToValueMap = map[string]AssignmentStatus{
	_AssignmentStatusName[0:23]:       AssignmentStatusINVALID,
	_AssignmentStatusLowerName[0:23]:  AssignmentStatusINVALID,
	_AssignmentStatusName[23:44]:      AssignmentStatusREADY,
	_AssignmentStatusLowerName[23:44]: AssignmentStatusREADY,
}

var _AssignmentStatusNames = []string{
	_AssignmentStatusName[0:23],
	_AssignmentStatusName[23:44],
}

// AssignmentStatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func AssignmentStatusString(s string) (AssignmentStatus, error) {
	if val, ok := _AssignmentStatusNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _AssignmentStatusNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to AssignmentStatus values", s)
}

// AssignmentStatusValues returns all values of the enum
func AssignmentStatusValues() []AssignmentStatus {
	return _AssignmentStatusValues
}

// AssignmentStatusStrings returns a slice of all String values of the enum
func AssignmentStatusStrings() []string {
	strs := make([]string, len(_AssignmentStatusNames))
	copy(strs, _AssignmentStatusNames)
	return strs
}

// IsAAssignmentStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i AssignmentStatus) IsAAssignmentStatus() bool {
	for _, v := range _AssignmentStatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for AssignmentStatus
func (i AssignmentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for AssignmentStatus
func (i *AssignmentStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AssignmentStatus should be a string, got %s", data)
	}

	var err error
	*i, err = AssignmentStatusString(s)
	return err
}
