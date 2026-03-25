package cli

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
)

func isJSON(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("json")
	return v
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
