package catalog

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/jsaevecke/k8s-system-labs/internal/domain"
)

type Labs []domain.Lab

func (l Labs) Print(output io.Writer) error {
	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "NAME\tCLUSTER PROVIDER\tLAB PROVIDER\tWORKERS"); err != nil {
		return fmt.Errorf("write lab list: %w", err)
	}

	for _, definition := range l {
		if _, err := fmt.Fprintf(
			table,
			"%s\t%s\t%s\t%d\n",
			definition.Metadata.Name,
			definition.Spec.Providers.Cluster.String(),
			definition.Spec.Providers.Lab.String(),
			definition.Spec.Cluster.Workers,
		); err != nil {
			return fmt.Errorf("write lab list: %w", err)
		}
	}

	if err := table.Flush(); err != nil {
		return fmt.Errorf("write lab list: %w", err)
	}
	return nil
}
