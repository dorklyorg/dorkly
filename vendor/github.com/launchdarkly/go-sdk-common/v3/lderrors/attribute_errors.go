package lderrors

// ErrAttributeEmpty means that you tried to use an uninitialized ldattr.Ref{}, or one that was initialized
// from an empty string, or from a string that consisted only of a slash.
//
// For details of the attribute reference syntax, see ldattr.Ref.
type ErrAttributeEmpty struct{}

// ErrAttributeExtraSlash means that an attribute reference contained a double slash or trailing slash
// causing one path component to be empty, such as "/a//b" or "/a/b/".
//
// For details of the attribute reference syntax, see ldattr.Ref.
type ErrAttributeExtraSlash struct{}

// ErrAttributeInvalidEscape means that an attribute reference contained contained a "~" character that was
// not followed by "0" or "1".
//
// For details of the attribute reference syntax, see ldattr.Ref.
type ErrAttributeInvalidEscape struct{}

const (
	msgAttributeEmpty         = "attribute reference cannot be empty"
	msgAttributeExtraSlash    = "attribute reference contained a double slash or a trailing slash"
	msgAttributeInvalidEscape = "attribute reference contained an escape character (~) that was not followed by 0 or 1"
)

func (e ErrAttributeEmpty) Error() string         { return msgAttributeEmpty }
func (e ErrAttributeExtraSlash) Error() string    { return msgAttributeExtraSlash }
func (e ErrAttributeInvalidEscape) Error() string { return msgAttributeInvalidEscape }
