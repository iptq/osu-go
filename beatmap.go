package osu

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Mode = int

const (
	MODE_STD   = 0
	MODE_TAIKO = 1
	MODE_CTB   = 2
	MODE_MANIA = 3
)

var (
	FILE_FORMAT_PATTERN = regexp.MustCompile(`^osu file format v(\d+)$`)
	SECTION_PATTERN     = regexp.MustCompile(`^\[([[:alpha:]]+)\]$`)
	KEY_VALUE_PATTERN   = regexp.MustCompile(`^([A-Za-z]+)\s*:\s*(.*)$`)
)

var WHAT_THE_FUCK = map[bool]int{false: 0, true: 1}

type Beatmap struct {
	Version int

	AudioFilename        string
	AudioLeadIn          int
	PreviewTime          int
	Countdown            bool
	SampleSet            SampleSet
	StackLeniency        float64
	Mode                 Mode
	LetterboxInBreaks    bool
	EpilepsyWarning      bool
	WidescreenStoryboard bool

	Title          string
	TitleUnicode   string
	Artist         string
	ArtistUnicode  string
	Creator        string
	DifficultyName string
	Source         string
	Tags           []string
	BeatmapID      int
	BeatmapSetID   int

	HPDrainRate       float64
	CircleSize        float64
	OverallDifficulty float64
	ApproachRate      float64
	SliderMultiplier  float64
	SliderTickRate    int

	Colors       []Color
	TimingPoints []*TimingPoint
	HitObjects   []*HitObject

	// TODO: events
}

func ParseBeatmap(reader io.Reader) (m *Beatmap, err error) {
	// Largely based on https://github.com/natsukagami/go-osu-parser/blob/master/parser.go
	var section string
	var buf []byte

	m = &Beatmap{}
	bufreader := bufio.NewReader(reader)

	// compatibility for older versions
	approachSet := false
	artistUnicodeSet := false
	titleUnicodeSet := false

	m.BeatmapSetID = -1

	for nLine := 0; err == nil; buf, _, err = bufreader.ReadLine() {
		nLine += 1

		line := strings.Trim(string(buf), " \r\n")
		if len(line) == 0 {
			// empty line
			continue
		}

		// check for osu file format header
		if match := FILE_FORMAT_PATTERN.FindStringSubmatch(line); match != nil {
			if n, err := strconv.Atoi(match[1]); err == nil {
				m.Version = n
			}
			continue
		}

		// update current section
		if match := SECTION_PATTERN.FindStringSubmatch(line); match != nil {
			section = match[1]
			continue
		}

		// yay all other sections
		switch strings.ToLower(section) {
		case "general":
			fallthrough
		case "editor":
			fallthrough
		case "metadata":
			fallthrough
		case "difficulty":
			if match := KEY_VALUE_PATTERN.FindStringSubmatch(line); match != nil {
				key, value := match[1], match[2]
				switch strings.ToLower(key) {
				// [General]
				case "audiofilename":
					// check that its extension is mp3
					if !strings.HasSuffix(strings.ToLower(value), ".mp3") {
						return nil, errors.New("AudioFilename does not have the .mp3 extension")
					}

					m.AudioFilename = value
				case "audioleadin":
					if val, err := strconv.Atoi(value); err == nil {
						m.AudioLeadIn = val
					}
				case "previewtime":
					if val, err := strconv.Atoi(value); err == nil {
						m.PreviewTime = val
					}
				case "countdown":
					if val, err := strconv.Atoi(value); err == nil {
						m.Countdown = val > 0
					}
				case "sampleset":
					m.SampleSet = SAMPLE_SETS_INV[strings.ToLower(value)]
				case "stackleniency":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						m.StackLeniency = val
					}
				case "letterboxinbreaks":
					if val, err := strconv.Atoi(value); err == nil {
						m.LetterboxInBreaks = val > 0
					}
				case "epilepsywarning":
					if val, err := strconv.Atoi(value); err == nil {
						m.EpilepsyWarning = val > 0
					}
				case "widescreenstoryboard":
					if val, err := strconv.Atoi(value); err == nil {
						m.WidescreenStoryboard = val > 0
					}

				// [Metadata]
				case "title":
					m.Title = value
				case "titleunicode":
					m.TitleUnicode = value
				case "artist":
					m.Artist = value
				case "artistunicode":
					m.ArtistUnicode = value
				case "creator":
					m.Creator = value
				case "version":
					m.DifficultyName = value
				case "source":
					m.Source = value
				case "tags":
					m.Tags = strings.Split(value, " ")
				case "beatmapid":
					if val, err := strconv.Atoi(value); err == nil {
						m.BeatmapID = val
					}
				case "beatmapsetid":
					if val, err := strconv.Atoi(value); err == nil {
						m.BeatmapSetID = val
					}

				// [Difficulty]
				case "hpdrainrate":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						m.HPDrainRate = val
					}
				case "circlesize":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						m.CircleSize = val
					}
				case "overalldifficulty":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						m.OverallDifficulty = val
					}
				case "approachrate":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						m.ApproachRate = val
					}

				default:
					// return nil, fmt.Errorf("unknown key '%s'", key)
				}
			} else {
				return nil, fmt.Errorf("line %d\tfailed to match: '%+v'", nLine, line)
			}
		case "events":
			// TODO:
		case "timingpoints":
			// TODO:
		case "colours":
			// TODO:
		case "hitobjects":
			if obj, err := ParseHitObject(line); err == nil {
				m.HitObjects = append(m.HitObjects, &obj)
			} else {
				return nil, fmt.Errorf("line %d\tinvalid hitobject: %s (line: '%s')", nLine, err, line)
			}
		default:
			return nil, fmt.Errorf("line %d\tunknown section '%s'", nLine, section)
		}
	}

	if err == io.EOF {
		// this is actually a success
		// set err to nil so we can return success
		err = nil
	}

	// compatibility for older versions
	if !approachSet {
		// AR used to be set by OD
		m.ApproachRate = m.OverallDifficulty
	}
	if !artistUnicodeSet {
		m.ArtistUnicode = m.Artist
	}
	if !titleUnicodeSet {
		m.TitleUnicode = m.Title
	}

	// return nil, fmt.Errorf("%#v", m)
	return
}

// Serialize renders the beatmap into
func (m *Beatmap) Serialize(writer io.Writer) (err error) {
	var line string

	fmt.Fprintf(writer, "osu file format v%d\n", m.Version)
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[General]\n")
	fmt.Fprintf(writer, "AudioFilename: %s\n", m.AudioFilename)
	fmt.Fprintf(writer, "AudioLeadIn: %d\n", m.AudioLeadIn)
	fmt.Fprintf(writer, "PreviewTime: %d\n", m.PreviewTime)
	fmt.Fprintf(writer, "Countdown: %d\n", WHAT_THE_FUCK[m.Countdown])
	fmt.Fprintf(writer, "SampleSet: %s\n", SAMPLE_SETS[m.SampleSet])
	fmt.Fprintf(writer, "StackLeniency: %f\n", m.StackLeniency)
	fmt.Fprintf(writer, "Mode: %d\n", m.Mode)
	fmt.Fprintf(writer, "LetterboxInBreaks: %d\n", WHAT_THE_FUCK[m.LetterboxInBreaks])
	fmt.Fprintf(writer, "EpilepsyWarning: %d\n", WHAT_THE_FUCK[m.EpilepsyWarning])
	fmt.Fprintf(writer, "WidescreenStoryboard: %d\n", WHAT_THE_FUCK[m.WidescreenStoryboard])
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[Metadata]\n")
	fmt.Fprintf(writer, "Title:%s\n", m.Title)
	fmt.Fprintf(writer, "TitleUnicode:%s\n", m.TitleUnicode)
	fmt.Fprintf(writer, "Artist:%s\n", m.Artist)
	fmt.Fprintf(writer, "ArtistUnicode:%s\n", m.ArtistUnicode)
	fmt.Fprintf(writer, "Creator:%s\n", m.Creator)
	fmt.Fprintf(writer, "Version:%s\n", m.DifficultyName)
	fmt.Fprintf(writer, "Source:%s\n", m.Source)
	fmt.Fprintf(writer, "Tags:%s\n", strings.Join(m.Tags, " "))
	fmt.Fprintf(writer, "BeatmapID:%d\n", m.BeatmapID)
	fmt.Fprintf(writer, "BeatmapSetID:%d\n", m.BeatmapSetID)
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[Difficulty]\n")
	fmt.Fprintf(writer, "HPDrainRate:%.01f\n", m.HPDrainRate)
	fmt.Fprintf(writer, "CircleSize:%.01f\n", m.CircleSize)
	fmt.Fprintf(writer, "OverallDifficulty:%.01f\n", m.OverallDifficulty)
	fmt.Fprintf(writer, "ApproachRate:%.01f\n", m.ApproachRate)
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[Events]\n")
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[TimingPoints]\n")
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[Colours]\n")
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "[HitObjects]\n")
	for _, obj := range m.HitObjects {
		// TODO: handle err
		line, err = (*obj).Serialize()
		if err != nil {
			return
		}

		fmt.Fprintf(writer, "%s\n", line)
	}
	fmt.Fprintf(writer, "\n")

	return
}
