package textfmt

import (
	"fmt"
	"strings"
)

const (
	treeLevelEmpty               = "        "
	treeLevelOngoing             = "  |     "
	treeLevelPrefix              = "  |%s: "
	treeLevelPrefixLast          = `  \%s: `
	treeLevelPrefixNamePadChar   = '-'
	treeLevelPrefixNamePadAmount = 3
)

func LeveledGraphString[N any](root N, getName func(N) string, getChildren func(N) []N) string {
	return leveledStr(getName, getChildren, "", "", root)
}

func makeTreeLevelPrefix(msg string) string {
	for len([]rune(msg)) < treeLevelPrefixNamePadAmount {
		msg = string(treeLevelPrefixNamePadChar) + msg
	}
	return fmt.Sprintf(treeLevelPrefix, msg)
}

func makeTreeLevelPrefixLast(msg string) string {
	for len([]rune(msg)) < treeLevelPrefixNamePadAmount {
		msg = string(treeLevelPrefixNamePadChar) + msg
	}
	return fmt.Sprintf(treeLevelPrefixLast, msg)
}

func leveledStr[N any](getName func(node N) string, getChildren func(node N) []N, firstPrefix, contPrefix string, n N) string {
	var sb strings.Builder

	sb.WriteString(firstPrefix)
	sb.WriteString(getName(n))

	children := getChildren(n)

	for i := range children {
		sb.WriteRune('\n')
		var leveledFirstPrefix string
		var leveledContPrefix string
		if i+1 < len(children) {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefix("")
			leveledContPrefix = contPrefix + treeLevelOngoing
		} else {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefixLast("")
			leveledContPrefix = contPrefix + treeLevelEmpty
		}
		itemOut := leveledStr(getName, getChildren, leveledFirstPrefix, leveledContPrefix, children[i])
		sb.WriteString(itemOut)
	}

	return sb.String()
}
