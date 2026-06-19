package collectors

import (
	"context"
	"regexp"
	"strings"
	"time"
)

var profileLinePattern = regexp.MustCompile(`^([\S]+)\s{2,}(.+?)\s{2,}(running|stopped|unknown|unavailable)\s{2,}`)

func CollectProfiles(ctx context.Context, runner Runner) SectionResult[ProfilesData] {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	output, err := runner.Run(ctx, "hermes", "profile", "list")
	if err != nil {
		return SectionResult[ProfilesData]{Error: errorString(err)}
	}

	data, err := parseProfiles(output)
	if err != nil {
		return SectionResult[ProfilesData]{Error: err.Error()}
	}
	return SectionResult[ProfilesData]{Data: &data}
}

func parseProfiles(output []byte) (ProfilesData, error) {
	lines := strings.Split(string(output), "\n")
	profiles := make([]Profile, 0)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "Profile") || strings.HasPrefix(line, "─") {
			continue
		}

		matches := profileLinePattern.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		nameField := matches[1]
		profiles = append(profiles, Profile{
			Name:    strings.TrimPrefix(nameField, "◆"),
			Model:   strings.TrimSpace(matches[2]),
			Gateway: strings.TrimSpace(matches[3]),
			Active:  strings.HasPrefix(nameField, "◆"),
		})
	}

	if len(profiles) == 0 {
		return ProfilesData{}, context.DeadlineExceeded // converted by caller? no
	}
	return ProfilesData{Profiles: profiles}, nil
}
