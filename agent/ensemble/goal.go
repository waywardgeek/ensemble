package ensemble

import "ensemble/internal/common"

// GoalConfig selects an Agent-owned file and opt-in, process-local idle reminders.
type GoalConfig = common.GoalConfig

// GoalStatus is current configuration/activity, not semantic task completion.
type GoalStatus = common.GoalStatus

// GoalController is the optional public configuration and status interface.
type GoalController = common.GoalController

// GoalChanged reports an accepted snapshot; replay reads GoalStatus directly.
type GoalChanged = common.GoalChanged
