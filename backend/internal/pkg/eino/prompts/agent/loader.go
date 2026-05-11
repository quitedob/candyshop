package agent

import _ "embed"

//go:embed trade_agent.md
var TradeAgentInstruction string

//go:embed assistant_base.md
var AssistantInstructionBase string

//go:embed compliance_addition.md
var AssistantInstructionComplianceAddition string
