package web

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/BakedSoups/goverment-job-portal-fixer/search_engine"
)

type IndexPage struct {
	PageTitle        string
	Query            string
	SelectedTags     []search_engine.Tag
	AvailableTags    []search_engine.Tag
	TagGroups        []search_engine.TagGroup
	MapDistricts     []MapDistrict
	MapPoints        []MapPoint
	SourceGroups     []SourceGroup
	SelectedTagIDs   string
	SelectedGovIDs   string
	SelectedGovCount int
	YOEMin           int
	YOEMax           int
	Count            int
	Results          []ResultView
	SearchParams     string
}

type ResultView struct {
	search_engine.Result
	SearchParams string
	ConceptTags  []CardTag
	MatchLabel   string
	MatchDetail  string
}

type SourceOption struct {
	ID       string
	Name     string
	Region   string
	Count    int
	Selected bool
}

type SourceGroup struct {
	Name          string
	Options       []SourceOption
	SelectedCount int
	AllSelected   bool
}

type MapDistrict struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Count    int    `json:"count"`
	Level    int    `json:"level"`
	Selected bool   `json:"selected"`
}

type MapPoint struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	Count     int     `json:"count"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CardTag struct {
	search_engine.Tag
	Matched bool
}

type JobPage struct {
	PageTitle    string
	Document     search_engine.Document
	SelectedTags []search_engine.Tag
	SignalTags   []search_engine.Tag
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	selectedTagIDs := parseTags(r.URL.Query().Get("tags"))
	selectedGovIDs := parseQueryValues(r.URL.Query()["gov"])
	results := s.engine.SearchTags(selectedTagIDs)
	if len(selectedTagIDs) == 0 {
		results = s.engine.Search(query)
	}
	results = filterByGovernments(results, selectedGovIDs)
	yoeMin, yoeMax := parseYOERange(r.URL.Query().Get("yoe_min"), r.URL.Query().Get("yoe_max"), r.URL.Query().Get("yoe"))
	results = filterByYOERange(results, yoeMin, yoeMax)

	selectedTags := make([]search_engine.Tag, 0, len(selectedTagIDs))
	for _, id := range selectedTagIDs {
		if tag, ok := s.engine.Tag(id); ok {
			selectedTags = append(selectedTags, tag)
		}
	}

	page := IndexPage{
		PageTitle:        "Bay Area Gov Jobs",
		Query:            query,
		SelectedTags:     selectedTags,
		AvailableTags:    s.engine.Tags(),
		TagGroups:        s.engine.TagGroups(),
		MapDistricts:     mapDistricts(results, selectedGovIDs),
		MapPoints:        mapPoints(results),
		SourceGroups:     s.sourceGroups(selectedGovIDs),
		SelectedTagIDs:   strings.Join(selectedTagIDs, ","),
		SelectedGovIDs:   strings.Join(selectedGovIDs, ","),
		SelectedGovCount: len(selectedGovIDs),
		YOEMin:           yoeMin,
		YOEMax:           yoeMax,
		Count:            len(results),
		Results:          s.resultViews(results, searchParams(query, selectedTagIDs, selectedGovIDs, yoeMin, yoeMax), activeTagIDs(query, selectedTagIDs, s.engine)),
		SearchParams:     searchParams(query, selectedTagIDs, selectedGovIDs, yoeMin, yoeMax),
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var sourceCoordinates = map[string][2]float64{
	"sf": {37.7749, -122.4194}, "contra-costa-county": {37.9161, -121.9977},
	"san-mateo-county": {37.4337, -122.4014}, "santa-clara-county": {37.3337, -121.8907},
	"marin-county": {38.0834, -122.7633}, "sonoma-county": {38.5780, -122.9888},
	"napa-county": {38.5025, -122.2654}, "solano-county": {38.3105, -121.9018},
	"alameda": {37.7652, -122.2416}, "berkeley": {37.8715, -122.2730},
	"fremont": {37.5485, -121.9886}, "hayward": {37.6688, -122.0808},
	"oakland": {37.8044, -122.2712}, "richmond": {37.9358, -122.3477},
	"mountain-view": {37.3861, -122.0839}, "palo-alto": {37.4419, -122.1430},
	"san-jose": {37.3382, -121.8863}, "santa-clara": {37.3541, -121.9552},
	"sunnyvale": {37.3688, -122.0363}, "vallejo": {38.1041, -122.2566},
}

func mapPoints(results []search_engine.Result) []MapPoint {
	points := map[string]MapPoint{}
	for _, result := range results {
		job := result.Document.Job
		coordinates, ok := sourceCoordinates[job.SourceID]
		if !ok {
			continue
		}
		point := points[job.SourceID]
		point.ID = job.SourceID
		point.Name = job.SourceName
		point.Region = job.SourceRegion
		point.Latitude = coordinates[0]
		point.Longitude = coordinates[1]
		point.Count++
		points[job.SourceID] = point
	}

	out := make([]MapPoint, 0, len(points))
	for _, point := range points {
		out = append(out, point)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Server) sourceGroups(selected []string) []SourceGroup {
	options := s.sourceOptions(selected)
	groups := make([]SourceGroup, 0, 5)
	for _, option := range options {
		if len(groups) == 0 || groups[len(groups)-1].Name != option.Region {
			groups = append(groups, SourceGroup{Name: option.Region})
		}
		last := len(groups) - 1
		groups[last].Options = append(groups[last].Options, option)
		if option.Selected {
			groups[last].SelectedCount++
		}
	}
	for i := range groups {
		groups[i].AllSelected = len(groups[i].Options) > 0 && groups[i].SelectedCount == len(groups[i].Options)
	}
	return groups
}

func (s *Server) sourceOptions(selected []string) []SourceOption {
	selectedSet := map[string]bool{}
	for _, id := range selected {
		selectedSet[id] = true
	}

	optionsByID := map[string]SourceOption{}
	for _, job := range s.jobs {
		if job.SourceID == "" {
			continue
		}
		option := optionsByID[job.SourceID]
		option.ID = job.SourceID
		option.Name = job.SourceName
		option.Region = job.SourceRegion
		option.Count++
		option.Selected = selectedSet[job.SourceID]
		optionsByID[job.SourceID] = option
	}

	options := make([]SourceOption, 0, len(optionsByID))
	for _, option := range optionsByID {
		options = append(options, option)
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].Region == options[j].Region {
			return options[i].Name < options[j].Name
		}
		return options[i].Region < options[j].Region
	})
	return options
}

func mapDistricts(results []search_engine.Result, selected []string) []MapDistrict {
	counts := regionCounts(results)
	selectedRegions := selectedRegionSet(selected)
	maxCount := maxRegionCount(counts)

	districts := []MapDistrict{
		{ID: "north-bay", Name: "North Bay", Count: counts["North Bay"]},
		{ID: "sf", Name: "SF", Count: counts["SF"]},
		{ID: "east-bay", Name: "East Bay", Count: counts["East Bay"]},
		{ID: "peninsula", Name: "Peninsula", Count: counts["Peninsula"]},
		{ID: "south-bay", Name: "South Bay", Count: counts["South Bay"]},
	}
	for i := range districts {
		districts[i].Level = heatLevel(districts[i].Count, maxCount)
		districts[i].Selected = selectedRegions[districts[i].Name]
	}
	return districts
}

func regionCounts(results []search_engine.Result) map[string]int {
	counts := map[string]int{}
	for _, result := range results {
		counts[result.Document.Job.SourceRegion]++
	}
	return counts
}

func maxRegionCount(counts map[string]int) int {
	maxCount := 0
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}
	return maxCount
}

func heatLevel(count int, maxCount int) int {
	if count == 0 || maxCount == 0 {
		return 0
	}
	if count*4 >= maxCount*3 {
		return 4
	}
	if count*2 >= maxCount {
		return 3
	}
	if count*4 >= maxCount {
		return 2
	}
	return 1
}

func selectedRegionSet(selected []string) map[string]bool {
	selectedSources := selectedSet(selected)
	regions := map[string]bool{}
	sourceRegions := map[string]string{
		"sf": "SF", "contra-costa-county": "East Bay", "san-mateo-county": "Peninsula",
		"santa-clara-county": "South Bay", "marin-county": "North Bay", "sonoma-county": "North Bay",
		"napa-county": "North Bay", "solano-county": "North Bay", "alameda": "East Bay",
		"berkeley": "East Bay", "fremont": "East Bay", "hayward": "East Bay", "oakland": "East Bay",
		"richmond": "East Bay", "mountain-view": "South Bay", "palo-alto": "Peninsula",
		"san-jose": "South Bay", "santa-clara": "South Bay", "sunnyvale": "South Bay", "vallejo": "North Bay",
	}
	for id := range selectedSources {
		if region := sourceRegions[id]; region != "" {
			regions[region] = true
		}
	}
	return regions
}

func selectedSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func filterByGovernments(results []search_engine.Result, selected []string) []search_engine.Result {
	if len(selected) == 0 {
		return results
	}
	allowed := map[string]bool{}
	for _, id := range selected {
		allowed[id] = true
	}
	filtered := make([]search_engine.Result, 0, len(results))
	for _, result := range results {
		if allowed[result.Document.Job.SourceID] {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func (s *Server) resultViews(results []search_engine.Result, params string, activeTags []string) []ResultView {
	out := make([]ResultView, 0, len(results))
	active := selectedSet(activeTags)
	for _, result := range results {
		tags := make([]CardTag, 0, len(result.Document.ConceptNames))
		for _, id := range result.Document.ConceptNames {
			if tag, ok := s.engine.Tag(id); ok {
				tags = append(tags, CardTag{Tag: tag, Matched: active[id]})
			}
		}
		sort.SliceStable(tags, func(i, j int) bool {
			if tags[i].Matched != tags[j].Matched {
				return tags[i].Matched
			}
			return tags[i].Label < tags[j].Label
		})
		label, detail := matchSummary(result, len(activeTags) > 0)
		out = append(out, ResultView{
			Result:       result,
			SearchParams: params,
			ConceptTags:  tags,
			MatchLabel:   label,
			MatchDetail:  detail,
		})
	}
	return out
}

func matchSummary(result search_engine.Result, hasCriteria bool) (string, string) {
	if !hasCriteria {
		return "", ""
	}
	for _, reason := range result.Reasons {
		if strings.Contains(reason, "matched title") {
			return "Title match", "Your search terms appear in the job title."
		}
	}
	if result.Score >= 30 {
		return "Strong match", "Several relevant skills or terms appear in the listing."
	}
	return "Related match", "Relevant skills or terms appear in the listing."
}

func activeTagIDs(query string, selected []string, engine *search_engine.Engine) []string {
	if len(selected) > 0 {
		return selected
	}
	return engine.CanonicalTerms(query)
}

func parseYOEBound(raw string, fallback int) int {
	yoe, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if yoe < 0 {
		yoe = 0
	}
	if yoe > 10 {
		yoe = 10
	}
	return yoe
}

func parseYOERange(minRaw string, maxRaw string, legacyMaxRaw string) (int, int) {
	if minRaw == "" && maxRaw == "" && legacyMaxRaw != "" {
		return 0, parseYOEBound(legacyMaxRaw, 10)
	}
	minYOE := parseYOEBound(minRaw, 0)
	maxYOE := parseYOEBound(maxRaw, 10)
	if minYOE > maxYOE {
		minYOE, maxYOE = maxYOE, minYOE
	}
	return minYOE, maxYOE
}

func filterByYOERange(results []search_engine.Result, minYOE int, maxYOE int) []search_engine.Result {
	filtered := make([]search_engine.Result, 0, len(results))
	for _, result := range results {
		job := result.Document.Job
		if !job.RequiredYOEFound {
			if minYOE == 0 {
				filtered = append(filtered, result)
			}
			continue
		}
		requiredMax := job.RequiredYOEMax
		if requiredMax < job.RequiredYOEMin {
			requiredMax = job.RequiredYOEMin
		}
		if job.RequiredYOEMin <= maxYOE && requiredMax >= minYOE {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func parseTags(raw string) []string {
	return parseQueryValues([]string{raw})
}

func parseQueryValues(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			out = append(out, part)
			seen[part] = true
		}
	}
	return out
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	doc, ok := s.engine.Document(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	selectedTagIDs := parseTags(r.URL.Query().Get("tags"))
	selectedTags := make([]search_engine.Tag, 0, len(selectedTagIDs))
	for _, id := range selectedTagIDs {
		if tag, ok := s.engine.Tag(id); ok {
			selectedTags = append(selectedTags, tag)
		}
	}

	page := JobPage{
		PageTitle:    doc.Job.Title,
		Document:     doc,
		SelectedTags: selectedTags,
		SignalTags:   s.signalTags(doc),
	}
	if err := s.templates.ExecuteTemplate(w, "job.html", page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) signalTags(doc search_engine.Document) []search_engine.Tag {
	tags := make([]search_engine.Tag, 0, len(doc.ConceptNames))
	for _, id := range doc.ConceptNames {
		if tag, ok := s.engine.Tag(id); ok {
			tags = append(tags, tag)
		}
	}
	return tags
}

func searchParams(query string, tags []string, govs []string, yoeMin int, yoeMax int) string {
	var parts []string
	if strings.TrimSpace(query) != "" {
		parts = append(parts, "q="+url.QueryEscape(query))
	}
	if len(tags) > 0 {
		parts = append(parts, "tags="+strings.Join(tags, ","))
	}
	if len(govs) > 0 {
		parts = append(parts, "gov="+strings.Join(govs, ","))
	}
	parts = append(parts, "yoe_min="+strconv.Itoa(yoeMin))
	parts = append(parts, "yoe_max="+strconv.Itoa(yoeMax))
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}
