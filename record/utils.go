package record

import (
	"fmt"
)

func getColorStringFromCohortPercent(percent float64) string {
	return fmt.Sprintf("rgba(62,127,187,%.2f)", percent/100)
}
