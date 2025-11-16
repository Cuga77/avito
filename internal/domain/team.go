package domain

type Team struct {
	ID      int           `json:"id" db:"id"`
	Name    string        `json:"name" db:"name"`
	Members []*TeamMember `json:"members,omitempty"`
}

func (t *Team) Validate() error {
	if t.Name == "" {
		return ErrInvalidInput
	}
	if t.Members == nil {
		return ErrInvalidInput
	}

	for _, member := range t.Members {
		if err := member.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Team) GetActiveMembers() []*TeamMember {
	var active []*TeamMember
	for _, member := range t.Members {
		if member.IsActive {
			active = append(active, member)
		}
	}
	return active
}

func (t *Team) GetActiveMembersExcluding(excludeUserIDs ...string) []*TeamMember {
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	var result []*TeamMember
	for _, member := range t.Members {
		if member.IsActive && !excludeMap[member.UserID] {
			result = append(result, member)
		}
	}
	return result
}

func (t *Team) HasMember(userID string) bool {
	for _, member := range t.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
