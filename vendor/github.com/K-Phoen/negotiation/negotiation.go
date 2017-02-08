package negotiation

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

var formats = map[string][]string{}
var reverseFormats = map[string]string{}

type Alternative struct {
	Name    string
	Value   string
	Quality float64
	Params  map[string]string

	mainType, subType string
}

// just to ease alternatives sorting
type alternativeList []Alternative

// Split Media tokens in type and sub-type parts
type MediaTokenizer func(string) (string, string, error)

// Evaluate default quality for catch-all values
type QualityEvaluator func(*Alternative)

func LanguageTokenizer(text string) (string, string, error) {
	typeInfos := strings.Split(text, "-")

	if len(typeInfos) == 2 {
		return typeInfos[0], typeInfos[1], nil
	}

	return typeInfos[0], "*", nil
}

func DefaultTokenizer(text string) (string, string, error) {
	typeInfos := strings.Split(text, "/")

	if len(typeInfos) != 2 {
		return "", "", errors.New("Impossible to parse media-type " + text)
	}

	return typeInfos[0], typeInfos[1], nil
}

func DefaultQualityEvaluator(alternative *Alternative) {
	if alternative.mainType == "*" {
		alternative.Quality = 0.01
	} else if alternative.subType == "*" {
		alternative.Quality = 0.02
	}
}

func LanguageQualityEvaluator(alternative *Alternative) {
	if alternative.mainType == "*" {
		alternative.Quality = 0.01
	}
}

func (alternatives alternativeList) Len() int {
	slice := []Alternative(alternatives)
	return len(slice)
}

func (alternatives alternativeList) Less(i, j int) bool {
	slice := []Alternative(alternatives)
	ai, aj := slice[i], slice[j]

	if ai.Quality > aj.Quality {
		return true
	}

	return false
}

func (alternatives alternativeList) Swap(i, j int) {
	slice := []Alternative(alternatives)
	slice[i], slice[j] = slice[j], slice[i]
}

func parseHeader(header string, mediaTokenizer MediaTokenizer, qualityEvaluator QualityEvaluator) (alternativeList, error) {
	parts := strings.Split(header, ",")
	var alternatives alternativeList = make([]Alternative, 0, len(parts))
	var err error

	for _, part := range parts {
		// create the new alternative
		a := Alternative{}
		a.Quality = 1.0
		a.Params = make(map[string]string)

		// find the value/media-range
		mediaRange := strings.Split(part, ";")
		a.Value = strings.Trim(mediaRange[0], " ")

		// parse its type and sub-type
		a.mainType, a.subType, err = mediaTokenizer(a.Value)
		if err != nil {
			return alternatives, err
		}

		// fix default quality for catch-all values
		qualityEvaluator(&a)

		// parse the parameters
		for _, param := range mediaRange[1:] {
			split := strings.SplitN(strings.Trim(param, " "), "=", 2)

			// ignore invalid splits
			if len(split) != 2 {
				continue
			}

			key := strings.Trim(split[0], " ")
			value := strings.Trim(split[1], " ")

			// the quality is stored in the parameters, so extract if from them
			if key == "q" {
				a.Quality, _ = strconv.ParseFloat(value, 64)
			} else {
				a.Params[key] = value
			}
		}

		alternatives = append(alternatives, a)
	}

	sort.Stable(alternatives)

	return alternatives, nil
}

func (a Alternative) matches(alternative string, mediaTokenizer MediaTokenizer) bool {
	if a.Value == alternative {
		return true
	}

	mainType, subType, err := mediaTokenizer(alternative)
	if err != nil {
		return false
	}

	if mainType == a.mainType && subType == "*" {
		return true
	}

	if mainType == a.mainType && a.subType == "*" {
		return true
	}

	if a.mainType == "*" && a.subType == "*" {
		return true
	}

	return false
}

func findMatch(alternatives []string, headerAlternative Alternative, mediaTokenizer MediaTokenizer) (string, bool) {
	for _, alternative := range alternatives {
		if headerAlternative.matches(alternative, mediaTokenizer) {
			return alternative, true
		}
	}

	return "", false
}

func expandFormats(acceptedFormats []string, formatsMap map[string][]string) []string {
	formats := make([]string, 0, len(acceptedFormats))

	for _, format := range acceptedFormats {
		mapping, ok := formatsMap[format]

		if ok {
			formats = append(formats, mapping...)
		} else {
			formats = append(formats, format)
		}
	}

	return formats
}

func RegisterFormat(name string, mimeTypes []string) {
	formats[name] = mimeTypes

	for _, mimeType := range mimeTypes {
		reverseFormats[mimeType] = name
	}
}

func Negotiate(header string, acceptedAlternatives []string, mediaTokenizer MediaTokenizer, qualityEvaluator QualityEvaluator) (*Alternative, error) {
	alternatives, err := parseHeader(header, mediaTokenizer, qualityEvaluator)

	if err != nil {
		return nil, err
	}

	if len(alternatives) == 0 {
		return nil, errors.New("No alternative could be extracted from the given header")
	}

	for _, header := range alternatives {
		alternative, ok := findMatch(acceptedAlternatives, header, mediaTokenizer)

		if ok {
			var name string

			// try to find if there is a custom format associated to the mime-type
			reverseFormat, ok := reverseFormats[alternative]
			if ok {
				name = reverseFormat
			}

			return &Alternative{
				Name:    name,
				Value:   alternative,
				Quality: header.Quality,
				Params:  header.Params,
			}, nil
		}
	}

	return nil, errors.New("No matching alternative found")
}

func NegotiateLanguage(header string, alternatives []string) (*Alternative, error) {
	return Negotiate(header, alternatives, LanguageTokenizer, LanguageQualityEvaluator)
}

func NegotiateAccept(header string, alternatives []string) (*Alternative, error) {
	alternatives = expandFormats(alternatives, formats)

	return Negotiate(header, alternatives, DefaultTokenizer, DefaultQualityEvaluator)
}
