package errors

const (
	// Error Messages - Casual tone throughout
	ErrInvalidMethod      = "Whoa whoa whoa there, counselor! '%s'? That's not gonna fly. We're talking GET, POST, PUT, DELETE - the classics!"
	ErrMissingURL         = "Listen pal, I can't work with nothing here! You gotta give me a URL - that's Internet Law 101!"
	ErrInvalidURL         = "That URL? Not gonna hold up in court, friend! Give me something that actually works!"
	ErrInvalidTimeout     = "Time(out) is money, friend! Give me actual seconds for timeout, not whatever that was supposed to be."
	ErrPresetNotFound     = "Here's the deal, sport - preset '%s' doesn't exist in my files! Do I look like a magician to you?"
	ErrKeyNotFound        = "Let me tell you something, gentlemen - key '%s' is nowhere in %s. Case closed!"
	ErrInvalidTarget      = "Hold up! '%s'? That's amateur hour! Stick to the real targets: body, headers/header, query, request, variables, filters - that's how we do business!"
	ErrPresetNameRequired = "Hey hey hey! Can't work magic without knowing which preset we're talking about here!(this one is good)"
	ErrTargetRequired     = "Trust me on this one - gonna need to specify body, headers, query, request, variables, or filters."
	ErrKeyValueRequired   = "Bottom line, chief - I need actual key=value pairs to work with, not thin air! (this one is great)"
	ErrInvalidKeyValue    = "Yeah, that key=value thing you got there? Not gonna cut it in my operation"
	ErrArgumentsNeeded    = "Between you and me, amigo - gonna need more 'arguments' than that to make this case"
	ErrUnsupportedMethod  = "Look, look, look - method '%s' isn't in my professional repertoire (idk if needs more explanation or conclusion but is okay)"
	ErrRequestBuildFailed = "Well well well, request building just fell apart - that's a technical breach of contract!"
	ErrHTTPRequestFailed  = "Here's the situation - network rejected the request! Money talks, but apparently not loud enough today!"
	ErrVariableLoadFailed = "Let me break this down for you - variables won't load! File system's pleading the fifth!"
	ErrVariableSaveFailed = "This is how it works - variable didn't save! Storage system's in contempt!"
	ErrFileLoadFailed     = "File '%s' went MIA, plain and simple - either moved or never existed! Case closed!"
	ErrFileSaveFailed     = "Here's what happened - can't save file '%s'! That's a permissions violation, no negotiation!"
	ErrDirectoryFailed    = "Directory access denied, counselor - file system's exercising its rights! End of story!"
	ErrReadlineSetup      = "Terminal's not accepting input - that's a critical system failure! We got a problem here!"
	ErrInputRead          = "Input went haywire on me - communication breakdown! Let's retry this case!"
	ErrEditorNotFound     = "No editor found in evidence! Set $EDITOR or install nano/vim - that's due process!"
	ErrEditorFailed       = "Editor crashed and burned - technical malpractice in progress: %v"
	ErrNoCurrentPreset    = "No active case on file! Use: saul [preset] [command] to open proceedings - that's the law!"
)

const (
	// Warning Messages
	WarnNoFiltersMatch = "Heads up champ - no fields matched filters %v, might wanna check that syntax"
	WarnPresetExists   = "Just so we're clear, pal - preset '%s' already exists! No harm, no foul!"
	WarnResponseLarge  = "That response is huge (%d bytes), even 'loco' maybe - giving you raw JSON instead of TOML! That's just good business!"
)
