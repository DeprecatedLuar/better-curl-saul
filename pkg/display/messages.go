// Package display provides centralized message constants and console output functions.
// All messages in Better-Curl-Saul use these constants to maintain the
// characteristic "Saul Goodman" personality throughout the application.
package display

const (
	// HTTP Request Configuration
	ErrInvalidMethod  = "Whoa whoa whoa there, counselor! '%s'? That's not gonna fly. We're talking GET, POST, PUT, DELETE - the classics!"
	ErrMissingURL     = "Listen pal, I can't work with nothing here! You gotta give me a URL - that's Internet Law 101!"
	ErrInvalidURL     = "That URL? Not gonna hold up in court, friend! Give me something that actually works!"
	ErrInvalidTimeout = "Time(out) is money, friend! Give me actual seconds for timeout, not whatever that was supposed to be."

	// Command & Preset Validation
	ErrPresetNotFound     = "Here's the deal, sport - preset '%s' doesn't exist in my files! Do I look like a magician to you?"
	ErrKeyNotFound        = "Let me tell you something, gentlemen - key '%s' is nowhere in %s. Case closed!"
	ErrInvalidTarget      = "Hold up! '%s'? That's amateur hour! Stick to the real targets: body, headers/header, query, request, variables, filters - that's how we do business!"
	ErrPresetNameRequired = "Hey hey hey! Can't work magic without knowing which preset we're talking about here!(this one is good)"
	ErrTargetRequired     = "Trust me on this one - gonna need to specify body, headers, query, request, variables, or filters."
	ErrKeyValueRequired   = "Bottom line, chief - I need actual key=value pairs to work with, not thin air! (this one is great)"
	ErrInvalidKeyValue    = "Yeah, that key=value thing you got there? Not gonna cut it in my operation"
	ErrArgumentsNeeded    = "Between you and me, amigo - gonna need more 'arguments' than that to make this case"

	// Request Execution
	ErrUnsupportedMethod  = "Look, look, look - method '%s' isn't in my professional repertoire (idk if needs more explanation or conclusion but is okay)"
	ErrRequestBuildFailed = "Well well well, request building just fell apart - that's a technical breach of contract!"
	ErrHTTPRequestFailed  = "Here's the situation - network rejected the request! Money talks, but apparently not loud enough today!"

	// Variable Operations
	ErrVariableLoadFailed = "Let me break this down for you - variables won't load! File system's pleading the fifth!"
	ErrVariableSaveFailed = "This is how it works - variable didn't save! Storage system's in contempt!"

	// File System Operations
	ErrFileLoadFailed  = "File '%s' went MIA, plain and simple - either moved or never existed! Case closed!"
	ErrFileSaveFailed  = "Here's what happened - can't save file '%s'! That's a permissions violation, no negotiation!"
	ErrDirectoryFailed = "Directory access denied, counselor - file system's exercising its rights! End of story!"

	// Terminal & Editor
	ErrReadlineSetup  = "Terminal's not accepting input - that's a critical system failure! We got a problem here!"
	ErrInputRead      = "Input went haywire on me - communication breakdown! Let's retry this case!"
	ErrEditorNotFound = "No editor found in evidence! Set $EDITOR or install nano/vim - that's due process!"
	ErrEditorFailed   = "Editor crashed and burned - technical malpractice in progress: %v"

	// Response Handling
	ErrNoCurrentPreset       = "No active case on file! Use: saul [preset] [command] to open proceedings - that's the law!"
	ErrFieldNameRequired     = "Listen, counselor - need to specify what field you want! Options are: body, headers, status, url, method, duration"
	ErrUnknownResponseField  = "That field '%s'? Not in my case files! Stick to the evidence: body, headers, status, url, method, duration"
	ErrResponseProcessFailed = "Response processing went sideways - technical difficulties in the evidence room: %v"
	ErrTempFileCreate        = "Can't create temporary file - system's not cooperating with me here! Technical difficulties!"
	ErrTempFileRead          = "Temp file's playing hard to get - can't read it! Something went sideways!"

	// Curl Import
	ErrEmptyCurlCommand = "Come on now, friend - you gave me an empty file! I need an actual curl command to work with!"
	ErrCurlParseFailed  = "That curl command's not holding up under scrutiny - syntax error, plain and simple: %v"
	ErrNoCurlURL        = "Listen pal, that curl command's missing the most important part - the URL! Can't make a case without an address!"

	// Session-related errors
	ErrSessionTTYFailed    = "Can't get terminal ID - that's a courtroom security breach: %v"
	ErrSessionConfigFailed = "Config path's playing hide and seek - can't locate it: %v"
	ErrSessionDirFailed    = "Session directory won't cooperate - creation failed: %v"

	// Variant-related errors
	ErrInvalidVariantPath    = "That variant path '%s'? Not gonna hold up - should be preset/variant format!"
	ErrVariantPresetMissing  = "Base preset '%s' is MIA - can't create variants from thin air!"
	ErrVariantsDirFailed     = "Variants directory won't budge - creation failure: %v"
	ErrVariantDirFailed      = "Variant directory hit a wall - can't create it: %v"
	ErrVariantMigrateFailed  = "File '%s' won't migrate - something's blocking it: %v"
	ErrVariantConfigFailed   = "Variant .config file won't write - that's a technical violation: %v"
	ErrVariantNotExist       = "variant '%s' does not exist"
	ErrVariantNotExistPreset = "Variant '%s' is nowhere to be found in preset '%s' - check the files, counselor!"
	ErrVariantDeleteFailed   = "Variant '%s' won't delete - system's putting up resistance: %v"
	ErrVariantsReadFailed    = "Variants directory won't open up - can't read the files: %v"

	// Command errors
	ErrCommandUnknownGlobal = "Unknown command '%s'? That's not in my playbook, friend!"
	ErrCommandUnknownPreset = "Preset command '%s'? Never heard of it - check the legal documents!"
	ErrPresetCreateFailed   = "Preset '%s' creation just crashed - technical difficulties: %v"
	ErrPresetNotFoundCreate = "Preset '%s' not found. Create with: saul create %s or saul %s --create - that's proper procedure!"
	ErrNoActivePreset       = "No active case on file! Need a preset name to work with - that's the rules!"

	// Copy command errors
	ErrCopyNeedsArgs      = "Listen pal, I can't copy thin air! Need both source AND destination - that's Contract Law 101!"
	ErrCopyDestExists     = "Whoa there, counselor! Destination preset '%s' already exists - can't duplicate the evidence!"
	ErrCopyInvalidVariant = "That variant path? Not gonna hold up in court, friend! Should be preset/variant format - that's proper procedure!"
	ErrCopySourceNotFound = "Here's the deal - source variant '%s' is MIA, nowhere in my files! Can't copy what doesn't exist!"
	ErrCopyPresetFailed   = "Preset copy just fell apart on me - technical difficulties in the filing room: %v"
	ErrCopyVariantFailed  = "Variant copy went sideways, amigo - something's blocking the transfer: %v"
	ErrCopyFileFailed     = "File '%s' won't cooperate with the copy operation - system's being difficult: %v"

	// History and response errors
	ErrNoHistory             = "No history found for preset '%s' - no previous cases on file, counselor!"
	ErrHistoryLoadFailed     = "History files won't open up - system malfunction: %v"
	ErrInvalidResponseNumber = "That response number '%s'? Not valid in my books - give me a real number or 'last'!"

	// Remove command errors
	ErrInvalidRemoveCommand = "That remove command? Not gonna fly - check the syntax, friend!"
	ErrRemoveTargetRequired = "Need to tell me WHAT to remove here - body, headers, query, request, or variables!"
	ErrNoTargetsRemoved     = "Nothing got removed - all targets were already empty! No harm done!"

	// List command errors
	ErrInvalidListCommand = "That list command '%s'? Not in my playbook - check the documentation!"
	ErrHomeDirFailed      = "Can't find home directory - system's lost its way: %v"

	// Parser errors
	ErrUnknownFlag           = "Unknown flag '%s' - that's not in the legal documents!"
	ErrVariantNameRequired   = "Variant name required for switch command - can't switch to nothing!"
	ErrNoActivePresetVariant = "No active preset for relative variant path - need a base case first!"

	// Status and other errors
	ErrRequestConfigFailed = "Request config won't load - technical breakdown: %v"
	ErrCurlImportNotImpl   = "Curl import via editor not yet implemented in refactored code - coming soon, counselor!"
)

const (
	// Warning Messages
	WarnNoFiltersMatch    = "Heads up champ - no fields matched filters %v, might wanna check that syntax"
	WarnPresetExists      = "Just so we're clear, pal - preset '%s' already exists! No harm, no foul!"
	WarnResponseLarge     = "That response is huge (%d bytes), even 'loco' maybe - giving you raw JSON instead of TOML! That's just good business!"
	WarnHistoryFailed     = "Listen, buddy - couldn't save that response to history! No biggie, but thought you should know!"
	WarnUpdateCheckFailed = "Listen friend, couldn't check for updates right now - network's being difficult! Try again later, no big deal!"
	WarnSessionSaveFailed = "Session save went sideways - no big deal, but you should know: %v"
	WarnVariantNotFound   = "Warning: variant '%s' from .config does not exist, using 'default'"
)

const (
	// Update Messages
	InfoUpdateAvailable = `Well well well, look what we got here - version %s is available! Time for an upgrade, champ!
Current: %s
Latest:  %s

Pick your poison based on how you installed Saul:
  curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/better-curl-saul/main/install.sh | bash
  eget DeprecatedLuar/better-curl-saul
  brew upgrade saul  # (Probably was not from here yet)`

	InfoUpToDate = "You're golden, amigo! Already running the latest and greatest version."
)
