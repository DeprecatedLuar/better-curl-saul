package errors

const (
	// Error Messages - Casual tone throughout
	ErrInvalidMethod      = "sorry champ \"%s\" isn't really a thing, but i'll let you try again"
	ErrMissingURL         = "listen pal, at least put in the URL. Come on"
	ErrInvalidURL         = "alright, so the \"U R L\" needs to start with one of these two here: 'http://' or 'https://'. Go get'em tiger"
	ErrInvalidTimeout     = "hey buddy, timeout's gotta be a number (seconds)"
	ErrPresetNotFound     = "whoops, preset '%s' doesn't exist yet chief"
	ErrKeyNotFound        = "can't find key '%s' in %s, friend"
	ErrInvalidTarget      = "nice try pal, but '%s' isn't valid. Try: body, headers/header, query, request, variables, filters"
	ErrPresetNameRequired = "buddy, you're gonna need to tell me which preset you want"
	ErrTargetRequired     = "come on chief, gotta specify what you want to work with (body, headers, query, request, variables, filters)"
	ErrKeyValueRequired   = "hey there, need some key=value pairs to work with"
	ErrInvalidKeyValue    = "that key=value format doesn't look right, try again sport"
	ErrArgumentsNeeded    = "you gonna need more arguments than that buddy (no pressure)"
	ErrUnsupportedMethod  = "sorry pal, don't know how to handle method '%s'"
	ErrRequestBuildFailed = "couldn't build that request for you, something went not as expected"
	ErrHTTPRequestFailed  = "request didn't go through, network's being difficult"
	ErrVariableLoadFailed = "had trouble loading your variables, chief"
	ErrVariableSaveFailed = "couldn't save that variable for ya, sorry about that"
	ErrFileLoadFailed     = "couldn't load file '%s', might not be there"
	ErrFileSaveFailed     = "couldn't save file '%s', permissions maybe?"
	ErrDirectoryFailed    = "had trouble with directory stuff, file system's being picky"
	ErrReadlineSetup      = "couldn't set up the text input interface, something's wonky"
	ErrInputRead          = "had trouble reading your input, try that again buddy"
	ErrEditorNotFound     = "no editor found. Please set $EDITOR environment variable or install nano/vim"
	ErrEditorFailed       = "editor didn't work out, something went sideways: %v"
)

const (
	// Warning Messages
	WarnNoFiltersMatch = "heads up - no fields matched filters %v, might wanna check that syntax"
	WarnPresetExists   = "preset '%s' already exists, friendly reminder, alright?"
	WarnResponseLarge  = "response's pretty big (%d bytes) - I won't parse it for you. Here is the JSON instead of TOML"
)
