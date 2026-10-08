package llm

import "example.com/ensemble/internal/common"

func (a *Actor) InspectSkills() (common.SkillInspection, error) {
	r, err := a.ask(common.ActorMessage{Kind: "inspect_skills"})
	return r.SkillInspection, err
}
func (a *Actor) ChangeSkill(op common.SkillOperation) (common.SkillResult, error) {
	r, err := a.ask(common.ActorMessage{Kind: "change_skill", SkillOperation: op})
	return r.SkillResult, err
}
func (a *Actor) receiveSkills(m common.ActorMessage) bool {
	reply := common.ActorReply{}
	switch m.Kind {
	case "inspect_skills":
		reply.SkillInspection = a.parent.SkillView()
	case "change_skill":
		reply.SkillResult, reply.Error = a.parent.ChangeSkill(m.SkillOperation)
		if reply.Error != nil {
			if _, typed := reply.Error.(*common.SkillError); !typed {
				a.persistence(reply.Error)
			}
		}
	default:
		return false
	}
	m.Reply <- reply
	return true
}
