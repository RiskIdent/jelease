package main

import "testing"

func TestNewJiraIssueSearchQuery(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		project     string
		customField uint
		want        string
	}{
		{
			name:        "no custom field",
			status:      "Grooming",
			project:     "platform/jelease",
			customField: 0,
			want:        `status = "Grooming" and labels = "platform/jelease"`,
		},
		{
			name:        "with custom field",
			status:      "Grooming",
			project:     "platform/jelease",
			customField: 12500,
			want:        `status = "Grooming" and (labels = "platform/jelease" or cf[12500] ~ "platform/jelease")`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := newJiraIssueSearchQuery(tc.status, tc.project, tc.customField)
			if tc.want != got {
				t.Errorf("Wrong query.\nwant: `%s`\ngot:  `%s`", tc.want, got)
			}
		})
	}
}
