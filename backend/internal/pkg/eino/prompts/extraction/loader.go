package extraction

import (
	_ "embed"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

//go:embed classification_system.md
var classificationSystem string

//go:embed classification_user.md
var classificationUser string

//go:embed extraction_system.md
var extractionSystem string

//go:embed extraction_user.md
var extractionUser string

//go:embed validation_system.md
var validationSystem string

//go:embed validation_user.md
var validationUser string

// DocumentClassificationPrompt identifies the type of trade document from raw text.
var DocumentClassificationPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(classificationSystem),
	schema.UserMessage(classificationUser),
)

// ExtractionPrompt extracts structured data from trade documents.
var ExtractionPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(extractionSystem),
	schema.UserMessage(extractionUser),
)

// ValidationPrompt checks consistency across multiple documents.
var ValidationPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(validationSystem),
	schema.UserMessage(validationUser),
)
