package link

import (
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker"
	"github.com/newrelic/release-toolkit/src/changelog/linker/mapper"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	dictionaryPathFlag          = "dictionary"
	sampleFlag                  = "sample"
	disableGithubValidationFlag = "disable-github-validation"
	chFilePermissions           = os.FileMode(0o666)
)

// Cmd is the cli.Command object for the link-dependencies command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:      "link-dependencies",
	Usage:     "Attempts to add links to the original changelogs for dependency bumps in changelog.yaml. The link is computed automatically when the dependency name is a full route or it's got from a dictionary file when present.",
	UsageText: `Link dependencies retrieves the links for each dependency detecting the link if the name is a full route or matching an entry in the dictionary file.`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    dictionaryPathFlag,
			EnvVars: common.EnvFor(dictionaryPathFlag),
			Usage: "Path to a dictionary file mapping dependencies to their changelogs. " +
				"A dictionary is a YAML file with a root dictionary object, which contains a map from " +
				"dependency names to a template that will be rendered into a URL pointing to its changelog. " +
				"The template link must be in Go tpl format and typically will include the {{.To.Original}} variable " +
				"that will be replaced by the last bumped version (execute link-dependencies with --sample flag to see a dictionary.yml sample)",
			Value: "",
		},
		&cli.BoolFlag{
			Name:    sampleFlag,
			EnvVars: common.EnvFor(sampleFlag),
			Usage:   "Prints a sample dictionary to stdout.",
		},
		&cli.BoolFlag{
			Name:    disableGithubValidationFlag,
			EnvVars: common.EnvFor(disableGithubValidationFlag),
			Usage: "Disables changelog links validation for automatically detected Github repositories. " +
				"Github links validation performs a request to the rendered link in order to check if it actually exits. It the validation " +
				"fails, it will try a new link with/without the version's leading 'v' (which is a common issue when rendering Github links). " +
				"If generating a valid link is not possible, no link will be obtained for that particular dependency. " +
				"When disabled, changelog links for Github repositories are directly rendered using " +
				"https://github.com/<org>/<repo>/releases/tag/<new-version> with no validation, so no external request are performed.",
			Value: false,
		},
	},
	Action: Link,
}

// Link is a command function which tries to add a link to the changelog of each dependency in a changelog
// computing them from each of the defined mappers.
//
//nolint:gocyclo,cyclop
func Link(cCtx *cli.Context) error {
	chPath := cCtx.String(common.YAMLFlag)

	if cCtx.Bool(sampleFlag) {
		sampleDic, err := sampleDictionary()
		if err != nil {
			return fmt.Errorf("generating sample dicctionary: %w", err)
		}
		_, _ = fmt.Fprintf(cCtx.App.Writer, "%s", string(sampleDic))
		return nil
	}

	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", chPath, err)
	}

	ch := &changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}
	chFile.Close()

	mappers := make([]linker.Mapper, 0)

	if dicPath := cCtx.String(dictionaryPathFlag); dicPath != "" {
		dicFile, errPath := os.Open(dicPath)
		if errPath != nil {
			return fmt.Errorf("opening linker dictionary  %q: %w", dicPath, errPath)
		}
		dic, errPath := mapper.NewDictionary(dicFile)
		if errPath != nil {
			return fmt.Errorf("creating validator: %w", errPath)
		}
		mappers = append(mappers, dic)
	}

	var githubMapper linker.Mapper = mapper.Github{}

	if !cCtx.Bool(disableGithubValidationFlag) {
		githubMapper = mapper.NewWithLeadingVCheck(githubMapper)
	}

	mappers = append(mappers, githubMapper)

	link := linker.New(mappers...)
	err = link.Link(ch)
	if err != nil {
		return fmt.Errorf("linking dependency changelogs: %w", err)
	}

	chFile, err = os.OpenFile(chPath, os.O_RDWR|os.O_TRUNC, chFilePermissions)
	if err != nil {
		return fmt.Errorf("truncating changelog file: %w", err)
	}
	defer chFile.Close()

	err = yaml.NewEncoder(chFile).Encode(ch)
	if err != nil {
		return fmt.Errorf("writing changelog to %q: %w", chPath, err)
	}

	return nil
}

//nolint:wrapcheck
func sampleDictionary() ([]byte, error) {
	sampleDictionary := mapper.Dictionary{
		Changelogs: map[string]string{
			"newrelic-infrastructure": "https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-{{.To.Original}}",
			"golangci-lint":           "https://github.com/golangci/golangci-lint/releases/tag/{{.To.Original}}",
		},
	}
	return yaml.Marshal(sampleDictionary)
}
