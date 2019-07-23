package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/qri-io/ioes"
	"github.com/qri-io/qri/fsi"
	"github.com/qri-io/qri/lib"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates a `qri status` command that compares working directory to prev version
func NewStatusCommand(f Factory, ioStreams ioes.IOStreams) *cobra.Command {
	o := &StatusOptions{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show status of working directory",
		Long:    ``,
		Example: ``,
		Annotations: map[string]string{
			"group": "dataset",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			if !o.Refs.IsLinked() {
				return o.RunForeign()
			}
			return o.Run()
		},
	}

	return cmd
}

// StatusOptions encapsulates state for the Status command
type StatusOptions struct {
	ioes.IOStreams

	Refs *RefSelect

	FSIMethods *lib.FSIMethods
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *StatusOptions) Complete(f Factory, args []string) (err error) {
	o.Refs, err = GetCurrentRefSelect(f, args, 1)
	if err != nil {
		return err
	}

	o.FSIMethods, err = f.FSIMethods()
	return
}

// Run executes the status command
func (o *StatusOptions) Run() (err error) {
	printRefSelect(o.ErrOut, o.Refs)

	res := []lib.StatusItem{}
	dir := o.Refs.Dir()
	if err := o.FSIMethods.Status(&dir, &res); err != nil {
		printErr(o.ErrOut, err)
		return nil
	}

	clean := true
	valid := true
	for _, si := range res {
		switch si.Type {
		case fsi.STRemoved:
			printErr(o.Out, fmt.Errorf("  %s:  %s", si.Type, si.Component))
			clean = false
		case fsi.STUnmodified:
			// noop
		case fsi.STAdd, fsi.STChange:
			printErr(o.Out, fmt.Errorf("  %s: %s (source: %s)", si.Type, si.Component, filepath.Base(si.SourceFile)))
			clean = false
		}
		// TODO(dlong): Validate each file / component, set `valid` to false if any problems exist
	}

	if clean {
		printSuccess(o.Out, "working directory clean")
	} else if valid {
		printSuccess(o.Out, "\nrun `qri save` to commit this dataset")
	}
	return nil
}

// RunForeign returns status on a non-fsi-linked reference
func (o *StatusOptions) RunForeign() (err error) {
	printInfo(o.ErrOut, "using foreign dataset reference: %s", o.Refs.Ref())
	printRefSelect(o.ErrOut, o.Refs)

	res := []lib.StatusItem{}
	ref := o.Refs.Ref()
	if err := o.FSIMethods.StoredStatus(&ref, &res); err != nil {
		printErr(o.ErrOut, err)
		return nil
	}

	for _, si := range res {
		printInfo(o.Out, fmt.Sprintf("  %s", si.Component))
	}

	return nil
}
