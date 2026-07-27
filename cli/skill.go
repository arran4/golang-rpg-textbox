package cli

import (
	"github.com/arran4/golang-rpg-textbox/cli/skill"
)

// SkillInstall is a subcommand `rpgtextbox skill install`
//
// Flags:
//
//	scope: --scope (default: "user") Installation scope
//	agent: --agent (default: "") Target agent
func SkillInstall(source string, scope string, agent string, args ...string) error {
	return skill.Install(source, scope, agent, args)
}

// SkillUpdate is a subcommand `rpgtextbox skill update`
//
// Flags:
//
//	all: --all (default: "false") Update all
//	force: --force (default: "false") Force update
func SkillUpdate(name string, all bool, force bool) error {
	return skill.Update(name, all, force)
}

// SkillRemove is a subcommand `rpgtextbox skill remove`
//
// Flags:
//
//	force: --force (default: "false") Force remove
func SkillRemove(name string, force bool) error {
	return skill.Remove(name, force)
}

// SkillList is a subcommand `rpgtextbox skill list`
//
// Flags:
//
//	format: --format (default: "text") Format of the output
func SkillList(format string) error {
	return skill.List(format)
}

// SkillInspect is a subcommand `rpgtextbox skill inspect`
func SkillInspect(name string) error {
	return skill.Inspect(name)
}
