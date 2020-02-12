package tss

import (
	"fmt"
	"strings"
	"time"
)

type timeoutError struct {
	timeout   time.Duration
	stage     string
	memberIDs []MemberID
}

func (t timeoutError) Error() string {
	if len(t.memberIDs) > 0 {
		stringIDs := []string{}

		for _, memberID := range t.memberIDs {
			stringIDs = append(stringIDs, memberID.String())
		}

		return fmt.Sprintf(
			"timeout [%s] exceeded on stage [%s] - still waiting for members: [%s]",
			t.timeout,
			t.stage,
			strings.Join(stringIDs, ", "),
		)
	}

	return fmt.Sprintf(
		"timeout [%s] exceeded on stage [%s]",
		t.timeout,
		t.stage,
	)
}
