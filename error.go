package engle

import "fmt"

type ConvertError struct {
	path       string
	kindString string
	error
}

func NewConvertError(path string, kindString string) error {
	return &ConvertError{
		path:       path,
		kindString: kindString,
	}
}

func (c *ConvertError) Error() string {
	return fmt.Sprintf("%s convert %s error", c.path, c.kindString)
}
