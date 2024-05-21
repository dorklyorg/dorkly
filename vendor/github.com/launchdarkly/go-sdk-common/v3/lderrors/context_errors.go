package lderrors

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

const (
	msgContextUninitialized          = "tried to use uninitialized Context"
	msgContextKeyEmpty               = "context key must not be empty"
	msgContextKeyNull                = "context key must not be null"
	msgContextKeyMissing             = `"key" property not found in JSON context object`
	msgContextKindEmpty              = "context kind cannot be empty"
	msgContextKindCannotBeKind       = `"kind" is not a valid context kind`
	msgContextKindMultiForSingleKind = `single context cannot have the kind "multi"`
	msgContextKindMultiWithNoKinds   = "multi-context must contain at least one kind"
	msgContextKindMultiDuplicates    = "multi-context cannot have same kind more than once"
	msgContextKindInvalidChars       = "context kind contains disallowed characters"
)

// ErrContextUninitialized means that you have tried to use an empty ldcontext.Context{} struct.
type ErrContextUninitialized struct{}

// ErrContextKeyEmpty means that the ldcontext.Context Key field was set to an empty string.
type ErrContextKeyEmpty struct{}

// ErrContextKeyNull means that the "key" property in the JSON representation of an ldcontext.Context
// had a null value. This is specific to JSON unmarshaling, since there is no way to specify a null
// value for this field programmatically.
type ErrContextKeyNull struct{}

// ErrContextKeyMissing means that the JSON representation of an ldcontext.Context had no "key"
// property.
type ErrContextKeyMissing struct{}

// ErrContextKindEmpty means that the "kind" property in the JSON representation of an
// ldcontext.Context had an empty string value. This is specific to JSON unmarshaling, since if you
// are creating a Context programmatically, an empty string is automatically changed to
// ldcontext.DefaultKind.
type ErrContextKindEmpty struct{}

// ErrContextKindCannotBeKind means that you have tried to set the ldcontext.Context Kind field to
// the string "kind". The Context schema does not allow this.
type ErrContextKindCannotBeKind struct{}

// ErrContextKindMultiForSingleKind means that you have tried to set the ldcontext.Context Kind
// field to the string "multi" for a single context. The Context schema does not allow this.
type ErrContextKindMultiForSingleKind struct{}

// ErrContextKindMultiWithNoKinds means that you have used an ldcontext constructor or builder
// for a multi-context but you did not specify any individual Contexts in it.
type ErrContextKindMultiWithNoKinds struct{}

// ErrContextKindMultiDuplicates means that you have used an ldcontext constructor or builder
// for a multi-context and you specified more than one individual Context in it with the
// same kind.
type ErrContextKindMultiDuplicates struct{}

// ErrContextKindInvalidChars means that you have tried to set the ldcontext.Context Kind field to
// a string that contained disallowed characters.
type ErrContextKindInvalidChars struct{}

// ErrContextPerKindErrors means that a multi-context contained at least one kind where the
// individual Context was invalid. There may be a separate validation error for each kind.
type ErrContextPerKindErrors struct {
	// Errors is a map where each key is the context kind (as a string) and the value is the
	// validation error for that kind.
	Errors map[string]error
}

func (e ErrContextUninitialized) Error() string          { return msgContextUninitialized }
func (e ErrContextKeyEmpty) Error() string               { return msgContextKeyEmpty }
func (e ErrContextKeyNull) Error() string                { return msgContextKeyNull }
func (e ErrContextKeyMissing) Error() string             { return msgContextKeyMissing }
func (e ErrContextKindEmpty) Error() string              { return msgContextKindEmpty }
func (e ErrContextKindCannotBeKind) Error() string       { return msgContextKindCannotBeKind }
func (e ErrContextKindMultiForSingleKind) Error() string { return msgContextKindMultiForSingleKind }
func (e ErrContextKindMultiWithNoKinds) Error() string   { return msgContextKindMultiWithNoKinds }
func (e ErrContextKindMultiDuplicates) Error() string    { return msgContextKindMultiDuplicates }
func (e ErrContextKindInvalidChars) Error() string       { return msgContextKindInvalidChars }

func (e ErrContextPerKindErrors) Error() string {
	sortedKeys := maps.Keys(e.Errors)
	sort.Strings(sortedKeys)
	messages := make([]string, 0, len(e.Errors))
	for _, kind := range sortedKeys {
		messages = append(messages, fmt.Sprintf("(%s) %s", kind, e.Errors[kind]))
	}
	return strings.Join(messages, ", ")
}
