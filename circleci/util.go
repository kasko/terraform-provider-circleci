package circleci

import (
	"fmt"
)

func maskCircleCiSecret(value string) string {

	var take int

	switch len := len(value); len {
	case 1:
		take = len - 0
	case 2, 3:
		take = len - 1
	case 4, 5:
		take = len - 2
	case 6, 7:
		take = len - 3
	default:
		take = len - 4
	}

	return fmt.Sprintf("xxxx%s", value[take:])
}
