package app

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"fraunhofer/fkie/yapscan"
	"fraunhofer/fkie/yapscan/procIO"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/hillu/go-yara/v4"
	"github.com/sirupsen/logrus"
	"github.com/targodan/go-errors"
	"github.com/urfave/cli/v2"
)

const yaraRulesNamespace = ""

var onExit func()

func initAppAction(c *cli.Context) error {
	lvl, err := logrus.ParseLevel(c.String("log-level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	switch c.String("log-path") {
	case "-":
		logrus.SetOutput(os.Stdout)
	case "--":
		logrus.SetOutput(os.Stderr)
	default:
		logfile, err := os.OpenFile(c.String("log-path"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return errors.Errorf("could not open logfile for writing, reason: %w", err)
		}
		logrus.SetOutput(logfile)
		logrus.StandardLogger().ExitFunc = func(code int) {
			if onExit != nil {
				onExit()
			}
			os.Exit(code)
		}
		onExit = func() {
			logfile.Close()
		}
	}
	logrus.WithField("arguments", os.Args).Debug("Program started.")
	return nil
}

func listProcesses(c *cli.Context) error {
	err := initAppAction(c)
	if err != nil {
		return err
	}

	pids, err := procIO.GetRunningPIDs()
	if err != nil {
		return errors.Newf("could not enumerate PIDs, reason: %w", err)
	}

	for _, pid := range pids {
		fmt.Println(pid)
	}

	return nil
}

func filterFromArgs(c *cli.Context) (yapscan.MemorySegmentFilter, error) {
	var err error
	i := 0

	filters := make([]yapscan.MemorySegmentFilter, 8)

	filters[i], err = BuildFilterPermissions(c.String("filter-permissions"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-permissions\", reason: %w", err)
	}
	i += 1
	filters[i], err = BuildFilterPermissionsExact(c.StringSlice("filter-permissions-exact"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-permissions-exact\", reason: %w", err)
	}
	i += 1
	filters[i], err = BuildFilterType(c.StringSlice("filter-type"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-type\", reason: %w", err)
	}
	i += 1
	filters[i], err = BuildFilterState(c.StringSlice("filter-state"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-state\", reason: %w", err)
	}
	i += 1
	filters[i], err = BuildFilterSizeMax(c.String("filter-size-max"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-size-max\", reason: %w", err)
	}
	i += 1
	filters[i], err = BuildFilterSizeMin(c.String("filter-size-min"))
	if err != nil {
		return nil, errors.Errorf("invalid flag \"--filter-size-min\", reason: %w", err)
	}
	i += 1

	return yapscan.NewAndFilter(filters...), nil
}

func listMemory(c *cli.Context) error {
	err := initAppAction(c)
	if err != nil {
		return err
	}

	if c.NArg() != 1 {
		return errors.Newf("expected exactly one argument, got %d", c.NArg())
	}
	pid_, err := strconv.ParseUint(c.Args().Get(0), 10, 64)
	if err != nil {
		return errors.Newf("\"%s\" is not a pid", c.Args().Get(0))
	}
	pid := int(pid_)

	f, err := filterFromArgs(c)
	if err != nil {
		return err
	}

	proc, err := procIO.OpenProcess(pid)
	if err != nil {
		return errors.Newf("could not open process with pid %d, reason: %w", pid, err)
	}

	segments, err := proc.MemorySegments()
	if err != nil {
		return errors.Newf("could not enumerate memory segments of process %d, reason: %w", pid, err)
	}
	for _, seg := range segments {
		fRes := f.Filter(seg)
		if !fRes.Result {
			continue
		}

		format := "%19s %8s %3s %7s %7s %s\n"

		fmt.Printf(format, procIO.FormatMemorySegmentAddress(seg), humanize.Bytes(uint64(seg.Size)), seg.CurrentPermissions, seg.Type, seg.State, seg.FilePath)

		if c.Bool("list-subdivided") {
			for i, sseg := range seg.SubSegments {
				addr := procIO.FormatMemorySegmentAddress(sseg)
				if i+1 < len(seg.SubSegments) {
					addr = "├" + addr
				} else {
					addr = "└" + addr
				}

				fmt.Printf(format, addr, humanize.Bytes(uint64(sseg.Size)), sseg.CurrentPermissions, sseg.Type, sseg.State, sseg.FilePath)
			}
		}
	}

	return nil
}

func dumpMemory(c *cli.Context) error {
	err := initAppAction(c)
	if err != nil {
		return err
	}

	filter, err := filterFromArgs(c)
	if err != nil {
		return err
	}

	var dumper io.WriteCloser
	if c.Bool("raw") {
		dumper = os.Stdout
	} else {
		dumper = hex.Dumper(os.Stdout)
		defer dumper.Close()
	}

	if c.NArg() != 1 && c.NArg() != 2 {
		return errors.Newf("expected exactly one or two arguments, got %d", c.NArg())
	}
	pid_, err := strconv.ParseUint(c.Args().Get(0), 10, 64)
	if err != nil {
		return errors.Newf("\"%s\" is not a pid", c.Args().Get(0))
	}
	pid := int(pid_)

	var addr uintptr
	allSegments := c.NArg() < 2
	if !allSegments {
		_, err = fmt.Sscan(c.Args().Get(1), &addr)
		if err != nil {
			return errors.Newf("\"%s\" is not an address", c.Args().Get(1))
		}
	}

	proc, err := procIO.OpenProcess(pid)
	if err != nil {
		return errors.Newf("could not open process %d, reason: %w", pid, err)
	}

	baseSegments, err := proc.MemorySegments()
	if err != nil {
		return errors.Newf("could not retrieve memory segments of process %d, reason: %w", pid, err)
	}
	// Unpack segments
	segments := make([]*procIO.MemorySegmentInfo, 0, len(baseSegments))
	for _, seg := range baseSegments {
		if seg.SubSegments == nil || len(seg.SubSegments) == 0 {
			segments = append(segments, seg)
		} else {
			segments = append(segments, seg.SubSegments...)
		}
	}

	readContiguous := c.Int("contiguous")
	found := false
	for i, seg := range segments {
		if seg.BaseAddress == addr || allSegments {
			found = true
		}
		match := filter.Filter(seg)
		if allSegments && !match.Result {
			continue
		}
		if found {
			rdr, err := procIO.NewMemoryReader(proc, seg)
			if err != nil {
				return errors.Newf("could not read memory of process %d at address 0x%016X, reason %w", pid, seg.BaseAddress, err)
			}

			if c.Bool("store") {
				fname := fmt.Sprintf("%d_%s_0x%X.bin", pid, seg.CurrentPermissions.String(), seg.BaseAddress)
				path := path.Join(c.String("storage-dir"), fname)
				outfile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
				if err != nil {
					return errors.Newf("could not create dump file \"%s\", reason: %w", path, err)
				}
				_, err = io.Copy(outfile, rdr)
				outfile.Close()
				if err != nil {
					return errors.Newf("could not dump segment to file \"%s\", reason: %w", path, err)
				}
			} else {
				_, err = io.Copy(dumper, rdr)
				if err != nil {
					return errors.Newf("could not read memory of process %d at address 0x%016X, reason %w", pid, seg.BaseAddress, err)
				}
			}

			if !allSegments &&
				(readContiguous == 0 || (i+1 < len(segments) && segments[i+1].BaseAddress != seg.BaseAddress+seg.Size)) {
				// Next segment is not contiguous
				break
			}
			readContiguous--
		}
	}
	if !found {
		errors.Newf("process %d has no memory segment starting with address 0x%016X", pid, addr)
	}
	return nil
}

func askYesNoAlways(msg string) (yes bool, always bool) {
	var validAnswer bool
	for !validAnswer {
		fmt.Print(msg)
		fmt.Print(" (y/a/N): ")
		var line string
		fmt.Scanln(&line)

		switch strings.ToLower(strings.Trim(line, " \t\r\n")) {
		case "y":
			yes = true
			validAnswer = true
		case "":
			fallthrough
		case "n":
			validAnswer = true
		case "a":
			yes = true
			always = true
			validAnswer = true
		default:
			fmt.Println("Invalid answer.")
		}
	}
	return
}

func askYesNoAlwaysNever(msg string) (yes bool, always bool, never bool) {
	var validAnswer bool
	for !validAnswer {
		fmt.Print(msg)
		fmt.Print(" (y/a/N/never): ")
		var line string
		fmt.Scanln(&line)

		switch strings.ToLower(strings.Trim(line, " \t\r\n")) {
		case "y":
			yes = true
			validAnswer = true
		case "":
			fallthrough
		case "n":
			validAnswer = true
		case "a":
			yes = true
			always = true
			validAnswer = true
		case "never":
			never = true
			validAnswer = true
		default:
			fmt.Println("Invalid answer.")
		}
	}
	return
}

func scan(c *cli.Context) error {
	err := initAppAction(c)
	if err != nil {
		return err
	}

	f, err := filterFromArgs(c)
	if err != nil {
		return err
	}

	if c.NArg() == 0 && !c.Bool("all") {
		return errors.Newf("expected at least one argument or flag \"--all\", got zero")
	}

	var rules *yara.Rules
	err = func() error {
		rulesFile, err := os.OpenFile(c.String("rules"), os.O_RDONLY, 0644)
		if err != nil {
			return errors.Newf("could not open rules file, reason: %w", err)
		}
		defer rulesFile.Close()

		buff := make([]byte, 4)
		_, err = io.ReadFull(rulesFile, buff)
		if err != nil {
			return errors.Newf("could not read rules file, reason: %w", err)
		}
		rulesFile.Seek(0, io.SeekStart)

		if bytes.Equal(buff, []byte("YARA")) {
			logrus.Debug("Yara rules file contains compiled rules.")

			rules, err = yara.ReadRules(rulesFile)
			if err != nil {
				return errors.Newf("could not read rules file, reason: %w", err)
			}
		} else {
			logrus.Debug("Yara rules file needs to be compiled.")

			compiler, err := yara.NewCompiler()
			if err != nil {
				return errors.Newf("could not create yara compiler, reason: %w", err)
			}
			err = compiler.AddFile(rulesFile, yaraRulesNamespace)
			if err != nil {
				return errors.Newf("could not compile yara rules, reason: %w", err)
			}

			rules, err = compiler.GetRules()
			if err != nil {
				return errors.Newf("could not compile yara rules, reason: %w", err)
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}

	yaraScanner, err := yapscan.NewYaraMemoryScanner(rules)
	if err != nil {
		return errors.Newf("could not initialize yara scanner, reason: %w", err)
	}

	var pids []int
	if c.Bool("all") {
		pids, err = procIO.GetRunningPIDs()
		if err != nil {
			return errors.Newf("could not enumerate PIDs, reason: %w", err)
		}
	} else {
		pids = make([]int, c.NArg())
		for i := 0; i < c.NArg(); i += 1 {
			pids[i], err = strconv.Atoi(c.Args().Get(i))
			if err != nil {
				return errors.Newf("argument \"%s\" is not a pid: %w", c.Args().Get(i), err)
			}
		}
	}

	reporter := yapscan.NewProgressReporter(os.Stdout, yapscan.NewPrettyFormatter())
	if c.Bool("full-report") || c.Bool("store-dumps") {
		tmpDir := path.Join(os.TempDir(), "yapscan")
		fmt.Println("Full report temp dir: ", tmpDir)
		logrus.Debug("Full report temp dir: ", tmpDir)
		gatherRep, err := yapscan.NewGatheredAnalysisReporter(tmpDir)
		if err != nil {
			return errors.Errorf("could not initialize analysis reporter, reason: %w", err)
		}
		gatherRep.ZIP = gatherRep.SuggestZIPName()
		gatherRep.DeleteAfterZipping = !c.Bool("keep")
		fmt.Printf("Full report will be written to \"%s\".\n", gatherRep.ZIP)
		if c.Bool("store-dumps") {
			err = gatherRep.WithFileDumpStorage("dumps")
			if err != nil {
				return errors.Errorf("could not initialize analysis reporter, reason: %w", err)
			}
			gatherRep.ZIPPassword = c.String("password")
		}
		reporter = &yapscan.MultiReporter{
			Reporters: []yapscan.Reporter{
				reporter,
				gatherRep,
			},
		}
	}
	defer func() {
		err := reporter.Close()
		if err != nil {
			fmt.Println(err)
			logrus.WithError(err).Error("Error closing reporter.")
		}
	}()

	err = reporter.ReportSystemInfo()
	if err != nil {
		logrus.WithError(err).Error("Could not report on system infos.")
	}

	err = reporter.ReportRules(rules)
	if err != nil {
		logrus.WithError(err).Error("Could not report on yara rules.")
	}

	alwaysSuspend := c.Bool("force")
	alwaysDumpWithoutSuspend := false
	neverDumpWithoutSuspend := false

	for _, pid := range pids {
		if pid == os.Getpid() {
			// Don't scan yourself as that will cause unwanted matches.
			continue
		}

		proc, err := procIO.OpenProcess(pid)
		if err != nil {
			logrus.WithError(err).Errorf("could not open process %d for scanning", pid)
			continue
		}
		defer func() {
			if err := proc.Close(); err != nil {
				logrus.Error(err)
			}
		}()

		resume := func() {}
		if c.Bool("suspend") {
			var suspend bool
			if alwaysSuspend {
				suspend = true
			} else {
				suspend, alwaysSuspend = askYesNoAlways(fmt.Sprintf("Suspend process %d?", pid))
				if !suspend && !alwaysDumpWithoutSuspend && !neverDumpWithoutSuspend {
					var dump bool
					dump, alwaysDumpWithoutSuspend, neverDumpWithoutSuspend = askYesNoAlwaysNever("Scan anyway?")
					if !dump {
						continue
					}
				}
			}

			if suspend {
				err = proc.Suspend()
				if err != nil {
					fmt.Println("Could not suspend process: ", err)
					logrus.WithError(err).Errorf("could not suspend process %d", pid)
					continue
				}
				resume = func() {
					err := proc.Resume()
					if err != nil {
						fmt.Println("Could not resume process: ", err)
						logrus.WithError(err).Errorf("could not resume process %d", pid)
					}
				}
			} else {
				if neverDumpWithoutSuspend {
					continue
				}
			}
		}

		scanner := yapscan.NewProcessScanner(proc, f, yaraScanner)

		progress, err := scanner.Scan()
		if err != nil {
			logrus.WithError(err).Errorf("an error occurred during scanning of process %d", pid)
			resume()
			continue
		}
		err = reporter.ConsumeScanProgress(progress)
		if err != nil {
			logrus.WithError(err).Error("an error occurred during progress report, there may be no other output")
			resume()
			continue
		}
		resume()
	}

	return nil
}

func RunApp(args []string) {
	suspendFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "suspend",
			Aliases: []string{"s"},
			Usage:   "suspend the process before reading its memory",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "don't ask before suspending a process",
			Value:   false,
		},
	}

	segmentFilterFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "filter-permissions",
			Aliases: []string{"f-perm"},
			Usage:   "only consider segments with the given permissions or more, examples: \"rw\" includes segments with rw, rc and rwx",
		},
		&cli.StringSliceFlag{
			Name:    "filter-permissions-exact",
			Aliases: []string{"f-perm-e"},
			Usage:   "comma separated list of permissions to be considered, supported permissions: r, rw, rc, rwx, rcx",
		},
		&cli.StringSliceFlag{
			Name:    "filter-type",
			Aliases: []string{"f-type"},
			Usage:   "comma separated list of considered types, supported types: image, mapped, private",
		},
		&cli.StringSliceFlag{
			Name:    "filter-state",
			Aliases: []string{"f-state"},
			Usage:   "comma separated list of considered states, supported states: free, commit, reserve",
			Value:   cli.NewStringSlice("commit"),
		},
		&cli.StringFlag{
			Name:    "filter-size-max",
			Aliases: []string{"f-size-max"},
			Usage:   "maximum size of memory segments to be considered, can be absolute (e.g. \"1.5GB\"), percentage of total RAM (e.g. \"10%T\") or percentage of free RAM (e.g. \"10%F\")",
			Value:   "10%F",
		},
		&cli.StringFlag{
			Name:    "filter-size-min",
			Aliases: []string{"f-size-min"},
			Usage:   "minimum size of memory segments to be considered",
		},
	}

	app := &cli.App{
		Name:        "yapscan",
		HelpName:    "yapscan",
		Description: "A yara based scanner for files and process memory with some extras.",
		Version:     "0.1.0",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Luca Corbatto",
				Email: "luca.corbatto@fkie.fraunhofer.de",
			},
		},
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "one of [trace, debug, info, warn, error, fatal, panic]",
				Value:   "panic",
			},
			&cli.StringFlag{
				Name:  "log-path",
				Usage: "path to the logfile, or \"-\" for stdout, or \"--\" for stderr",
				Value: "--",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:    "list-processes",
				Aliases: []string{"ps", "lsproc"},
				Usage:   "lists all running processes",
				Action:  listProcesses,
			},
			&cli.Command{
				Name:      "list-process-memory",
				Aliases:   []string{"lsmem"},
				Usage:     "lists all memory segments of a process",
				ArgsUsage: "<pid>",
				Flags: append(append([]cli.Flag{
					&cli.BoolFlag{
						Name:  "list-free",
						Usage: "also list free memory segments",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "list-subdivided",
						Usage: "list segment subdivisions as they are now, as opposed to segments as they were allocated once",
					},
				}, segmentFilterFlags...), suspendFlags...),
				Action: listMemory,
			},
			&cli.Command{
				Name:      "dump",
				Usage:     "dumps memory of a process",
				Action:    dumpMemory,
				ArgsUsage: "<pid> [address_of_section]",
				Flags: append(append([]cli.Flag{
					&cli.IntFlag{
						Name:    "contiguous",
						Aliases: []string{"c"},
						Usage:   "also dump the following <value> contiguous sections, -1 for all contiguous sections, only relevant if [address_of_section] is given",
					},
					&cli.BoolFlag{
						Name:    "raw",
						Aliases: []string{"r"},
						Usage:   "dump the raw memory as opposed to a hex view of the memory",
						Value:   false,
					},
					&cli.BoolFlag{
						Name:  "store",
						Usage: "don't output, but store raw matching segments in --storage-dir",
						Value: false,
					},
					&cli.StringFlag{
						Name:    "storage-dir",
						Aliases: []string{"d"},
						Usage:   "directory for stored segments, ignored unless --store is given",
						Value:   ".",
					},
				}, suspendFlags...), segmentFilterFlags...),
			},
			&cli.Command{
				Name:      "scan",
				Usage:     "scans processes with yara rules",
				Action:    scan,
				ArgsUsage: "[pid...]",
				Flags: append(append([]cli.Flag{
					&cli.StringFlag{
						Name:     "rules",
						Aliases:  []string{"r", "C"},
						Usage:    "path to yara rules file, can be compiled or uncompiled",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "all",
						Usage: "scan all running processes",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "full-report",
						Usage: "create a full report",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "store-dumps",
						Usage: "store dumps of memory regions that match rules, implies --full-report, the report will be encrypted with --password",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "keep",
						Usage: "keep the temporary report directory, by default it will be deleted; ignored without --full-report",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "the password of the encrypted report, ignored unless --store-dumps is set",
						Value: yapscan.DefaultZIPPassword,
					},
				}, segmentFilterFlags...), suspendFlags...),
			},
		},
	}

	err := app.Run(args)
	if err != nil {
		fmt.Println(err)
		logrus.Error(err)
		logrus.Fatal("Aborting.")
	}
	if onExit != nil {
		onExit()
	}
}
