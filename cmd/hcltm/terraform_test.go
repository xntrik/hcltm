package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/xntrik/hcltm/pkg/spec"

	"github.com/kami-zh/go-capturer"
)

func TestTfRunEmpty(t *testing.T) {
	cmd := testTfCommand(t)

	var code int

	out := capturer.CaptureStdout(func() {
		code = cmd.Run([]string{})
	})

	if code != 1 {
		t.Errorf("Code did not equal 0: %d", code)
	}

	if !strings.Contains(out, "Please provide <files> or -stdin") {
		t.Errorf("Expected %s to contain %s", out, "Please provide <files> or -stdin")
	}
}

func TestTfRunNoFile(t *testing.T) {
	cmd := testTfCommand(t)

	var code int

	out := capturer.CaptureStdout(func() {
		code = cmd.Run([]string{"nofile"})
	})

	if code != 1 {
		t.Errorf("Code did not equal 0: %d", code)
	}

	if !strings.Contains(out, "no such file") {
		t.Errorf("Expected %s to contain %s", out, "no such file")
	}
}

func TestTfParsing(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		exp       []string
		invertexp bool
		code      int
		flags     string
	}{
		{
			"aws_s3_plan_broken",
			"./testdata/aws_s3/aws_s3.plan-broken-json",
			[]string{
				"Error unmarshalling JSON"},
			false,
			1,
			"-stdin",
		},
		{
			"aws_s3_plan_broken2_unknown_mode",
			"./testdata/aws_s3/aws_s3.plan-broken2-json",
			[]string{
				"Unknown mode"},
			false,
			1,
			"-stdin",
		},
		{
			"aws_s3_state_broken",
			"./testdata/aws_s3/aws_s3.state-broken-json",
			[]string{
				"Error unmarshalling JSON"},
			false,
			1,
			"-stdin",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := testTfCommand(t)

			var code int

			out := capturer.CaptureStdout(func() {
				var content []byte
				var err error

				if tc.in != "" {
					content, err = ioutil.ReadFile(tc.in)
					if err != nil {
						t.Fatal(err)
					}
				}
				tmpFile, err := ioutil.TempFile("", "example")
				if err != nil {
					t.Fatal(err)
				}

				defer os.Remove(tmpFile.Name())

				if _, err := tmpFile.Write(content); err != nil {
					t.Fatal(err)
				}

				if _, err := tmpFile.Seek(0, 0); err != nil {
					t.Fatal(err)
				}

				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()

				os.Stdin = tmpFile

				code = cmd.Run([]string{
					tc.flags,
					tc.in,
				})
			})

			if code != tc.code {
				t.Errorf("Code did not equal %d: %d", tc.code, code)
			}

			if !tc.invertexp {
				for _, exp := range tc.exp {
					if !strings.Contains(out, exp) {
						t.Errorf("Expected %s to contain %s", out, exp)
					}
				}
			} else {
				for _, exp := range tc.exp {
					if strings.Contains(out, exp) {
						t.Errorf("Was not expecting %s to contain %s", out, exp)
					}
				}
			}
		})
	}
}

func TestTfRunStdin(t *testing.T) {

	cases := []struct {
		name      string
		in        string
		exp       []string
		invertexp bool
		code      int
		flags     string
	}{
		{
			"aws_s3_plan",
			"./testdata/aws_s3/aws_s3.plan-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: my-tf-test-bucket",
				"terraform plan"},
			false,
			0,
			"-stdin",
		},
		{
			"aws_s3_state",
			"./testdata/aws_s3/aws_s3.state-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: xnmy-tf-test-bucket",
				"terraform state"},
			false,
			0,
			"-stdin",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := testTfCommand(t)

			var code int

			out := capturer.CaptureStdout(func() {
				var content []byte
				var err error

				if tc.in != "" {
					content, err = ioutil.ReadFile(tc.in)
					if err != nil {
						t.Fatal(err)
					}
				}
				tmpFile, err := ioutil.TempFile("", "example")
				if err != nil {
					t.Fatal(err)
				}

				defer os.Remove(tmpFile.Name())

				if _, err := tmpFile.Write(content); err != nil {
					t.Fatal(err)
				}

				if _, err := tmpFile.Seek(0, 0); err != nil {
					t.Fatal(err)
				}

				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()

				os.Stdin = tmpFile

				code = cmd.Run([]string{
					tc.flags,
					tc.in,
				})
			})

			if code != tc.code {
				t.Errorf("Code did not equal %d: %d", tc.code, code)
			}

			if !tc.invertexp {
				for _, exp := range tc.exp {
					if !strings.Contains(out, exp) {
						t.Errorf("Expected %s to contain %s", out, exp)
					}
				}
			} else {
				for _, exp := range tc.exp {
					if strings.Contains(out, exp) {
						t.Errorf("Was not expecting %s to contain %s", out, exp)
					}
				}
			}
		})
	}

}

func TestTfRun(t *testing.T) {

	cases := []struct {
		name      string
		in        string
		exp       []string
		invertexp bool
		code      int
		flags     []string
	}{
		{
			"aws_s3_plan",
			"./testdata/aws_s3/aws_s3.plan-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: my-tf-test-bucket",
				"terraform plan"},
			false,
			0,
			nil,
		},
		{
			"aws_s3_plan_default_class",
			"./testdata/aws_s3/aws_s3.plan-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: my-tf-test-bucket",
				"blep",
				"terraform plan"},
			false,
			0,
			[]string{"-default-classification=blep"},
		},
		{
			"aws_s3_state_default_class",
			"./testdata/aws_s3/aws_s3.state-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: xnmy-tf-test-bucket",
				"blep",
				"terraform state"},
			false,
			0,
			[]string{"-default-classification=blep"},
		},
		{
			"aws_s3_state",
			"./testdata/aws_s3/aws_s3.state-json",
			[]string{
				"information_asset \"aws_s3_bucket b\"",
				"bucket: xnmy-tf-test-bucket",
				"terraform state"},
			false,
			0,
			nil,
		},
		{
			"add_to_existing",
			"./testdata/aws_s3/aws_s3.plan-json",
			[]string{
				"This is some arbitrary text",
				"bucket: my-tf-test-bucket",
				"terraform plan",
			},
			false,
			0,
			[]string{"-add-to-existing=./testdata/tm3.hcl"},
		},
		{
			"add_to_existing_state",
			"./testdata/aws_s3/aws_s3.state-json",
			[]string{
				"This is some arbitrary text",
				"bucket: xnmy-tf-test-bucket",
				"terraform state",
			},
			false,
			0,
			[]string{"-add-to-existing=./testdata/tm3.hcl"},
		},
		{
			"add_to_existing_no_in",
			"",
			[]string{
				"contains multiple models",
				"tm1 one",
				"tm tm1 two",
			},
			false,
			1,
			[]string{"-add-to-existing=./testdata/tm1.hcl"},
		},
		{
			"add_to_existing_tm_select_invalid_no_in",
			"",
			[]string{
				"contains multiple models",
				"tm1 one",
				"tm tm1 two",
			},
			false,
			1,
			[]string{"-add-to-existing=./testdata/tm1.hcl", "-tm-name=b"},
		},
		{
			"add_to_existing_tm_select_valid_no_in",
			"",
			[]string{
				"Please provide <files> or -stdin",
			},
			false,
			1,
			[]string{"-tm-name=tm1 one", "-add-to-existing=./testdata/tm1.hcl"},
		},
		{
			"add_to_existing_tm_select_empty_hcl",
			"",
			[]string{
				"Need at least 1 threat model",
			},
			false,
			1,
			[]string{"-add-to-existing=./testdata/tm4.hcl"},
		},
		{
			"add_to_existing_tm_select_single_hcl",
			"",
			[]string{
				"Please provide <files> or -stdin",
			},
			false,
			1,
			[]string{"-add-to-existing=./testdata/tm3.hcl"},
		},
		{
			"add_to_existing_tm_select_broken_hcl",
			"",
			[]string{
				"Error parsing provided <hcltm file>:",
			},
			false,
			1,
			[]string{"-add-to-existing=./testdata/tm3-broken.hcl.test"},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := testTfCommand(t)

			input := []string{}
			if tc.flags != nil {
				for _, tcflag := range tc.flags {
					input = append(input, tcflag)
				}
			}
			if tc.in != "" {
				input = append(input, tc.in)
			}

			var code int

			out := capturer.CaptureStdout(func() {
				code = cmd.Run(input)
			})

			if code != tc.code {
				t.Errorf("Code did not equal %d: %d", tc.code, code)
			}

			if !tc.invertexp {
				for _, exp := range tc.exp {
					if !strings.Contains(out, exp) {
						t.Errorf("Expected %s to contain %s", out, exp)
					}
				}
			} else {
				for _, exp := range tc.exp {
					if strings.Contains(out, exp) {
						t.Errorf("Was not expecting %s to contain %s", out, exp)
					}
				}
			}
		})
	}

}

func testTfCommand(tb testing.TB) *TerraformCommand {
	tb.Helper()

	d, err := ioutil.TempDir("", "")
	if err != nil {
		tb.Fatalf("Error creating tmp dir: %s", err)
	}

	_ = os.Setenv("HOME", d)

	cfg, _ := spec.LoadSpecConfig()

	defer os.RemoveAll(d)

	global := &GlobalCmdOptions{}

	return &TerraformCommand{
		GlobalCmdOptions: global,
		specCfg:          cfg,
	}
}
