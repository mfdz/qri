package cmd

import (
	"strings"
	"testing"

	"github.com/qri-io/qri/errors"
)

func TestValidateComplete(t *testing.T) {
	run := NewTestRunner(t, "test_peer", "qri_test_validate_complete")
	defer run.Delete()

	f, err := NewTestFactory()
	if err != nil {
		t.Errorf("error creating new test factory: %s", err)
		return
	}

	cases := []struct {
		args           []string
		bodyFilepath   string
		schemaFilepath string
		expect         string
		err            string
	}{
		{[]string{}, "filepath", "schemafilepath", "", ""},
		{[]string{"test"}, "", "", "test", ""},
		{[]string{"foo", "bar"}, "", "", "foo", ""},
	}

	for i, c := range cases {
		opt := &ValidateOptions{
			IOStreams:      run.Streams,
			BodyFilepath:   c.bodyFilepath,
			SchemaFilepath: c.schemaFilepath,
		}
		opt.Complete(f, c.args)

		if c.err != run.ErrStream.String() {
			t.Errorf("case %d, error mismatch. Expected: '%s', Got: '%s'", i, c.err, run.ErrStream.String())
			run.IOReset()
			continue
		}

		if c.expect != opt.Refs.Ref() {
			t.Errorf("case %d, opt.Refs not set correctly. Expected: '%s', Got: '%s'", i, c.expect, opt.Refs.Ref())
			run.IOReset()
			continue
		}

		if opt.DatasetMethods == nil {
			t.Errorf("case %d, opt.DatasetMethods not set.", i)
			run.IOReset()
			continue
		}
		run.IOReset()
	}

}

func TestValidateRun(t *testing.T) {
	run := NewTestRunner(t, "test_peer", "qri_test_validate_complete")
	defer run.Delete()

	f, err := NewTestFactory()
	if err != nil {
		t.Errorf("error creating new test factory: %s", err)
		return
	}

	cases := []struct {
		description    string
		ref            string
		bodyFilePath   string
		schemaFilePath string
		url            string
		expected       string
		err            string
		msg            string
	}{
		{"bad args", "", "", "", "", "", "bad arguments provided", "please provide a dataset name, or a supply the --body and --schema or --structure flags"},
		// TODO: add back when we again support validating from a URL
		// {"", "", "", "url", "", "bad arguments provided", "if you are validating data from a url, please include a dataset name or supply the --schema flag with a file path that Qri can validate against"},
		{"movie problems", "peer/movies", "", "", "", movieOutput, "", ""},
		{"dataset not found", "peer/bad_dataset", "", "", "", "", "cannot find dataset: peer/bad_dataset", ""},
		{"body file not found", "", "bad/filepath", "testdata/days_of_week_schema.json", "", "", "error opening body file: bad/filepath", ""},
		{"schema file not found", "", "testdata/days_of_week.csv", "bad/schema_filepath", "", "", "error opening schema file: bad/schema_filepath", ""},
		{"validate successfully", "", "testdata/days_of_week.csv", "testdata/days_of_week_schema.json", "", "✔ All good!\n", "", ""},
		// TODO: pull from url
	}

	for i, c := range cases {
		dsm, err := f.DatasetMethods()
		if err != nil {
			t.Errorf("case %d, error creating dataset request: %s", i, err)
			continue
		}

		opt := &ValidateOptions{
			IOStreams:      run.Streams,
			Refs:           NewExplicitRefSelect(c.ref),
			BodyFilepath:   c.bodyFilePath,
			SchemaFilepath: c.schemaFilePath,
			URL:            c.url,
			DatasetMethods: dsm,
		}

		err = opt.Run()
		if (err == nil && c.err != "") || (err != nil && c.err != err.Error()) {
			t.Errorf("case %d, mismatched error. Expected: '%s', Got: '%v'", i, c.err, err)
			run.IOReset()
			continue
		}

		if libErr, ok := err.(errors.Error); ok {
			if libErr.Message() != c.msg {
				t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: '%s'", i, c.msg, libErr.Message())
				run.IOReset()
				continue
			}
		} else if c.msg != "" {
			t.Errorf("case %d, mismatched user-friendly message. Expected: '%s', Got: ''", i, c.msg)
			run.IOReset()
			continue
		}

		if c.expected != run.OutStream.String() {
			t.Errorf("case %d, output mismatch. Expected: '%s', Got: '%s'", i, c.expected, run.OutStream.String())
			run.IOReset()
			continue
		}

		run.IOReset()
	}
}

var movieOutput = `0: /4/1: "" type should be integer
1: /199/1: "" type should be integer
2: /206/1: "" type should be integer
3: /1510/1: "" type should be integer
`

func TestValidateCommandlineFlags(t *testing.T) {
	run := NewTestRunner(t, "test_peer", "qri_test_validate_commandline_flags")
	defer run.Delete()

	output := run.MustExec(t, "qri validate --body=testdata/movies/body_ten.csv --structure=testdata/movies/structure_override.json")
	expectContain := `/4/1: "" type should be integer`

	if !strings.Contains(output, expectContain) {
		t.Errorf("expected output to contain %q, got %q", expectContain, output)
	}

	output = run.MustExec(t, "qri validate --body=testdata/movies/body_ten.csv --schema=testdata/movies/schema_only.json")
	expectContain = `/0/1: "duration" type should be integer
1: /5/1: "" type should be integer`

	if !strings.Contains(output, expectContain) {
		t.Errorf("expected output to contain %q, got %q", expectContain, output)
	}

	// Fail because both --structure and --schema are given
	err := run.ExecCommand("qri validate --body=testdata/movies/body_ten.csv --structure=testdata/movies/structure_override.json --schema=testdata/movies/schema_only.json")
	if err == nil {
		t.Fatal("expected error, did not get one")
	}
	expect := "bad arguments provided"
	if expect != err.Error() {
		t.Errorf("expected %q, got %q", expect, err.Error())
	}
}
