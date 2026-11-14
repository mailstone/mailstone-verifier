package main

import _ "embed"

// The licence of MailStone Verifier and the notices of the third-party
// components compiled into it (scripts/mentions-tiers.sh), shown in About.

//go:embed LICENSE
var licenseText string

//go:embed THIRD-PARTY-NOTICES.md
var noticesText string

// GetLicense returns the licence of MailStone Verifier.
func (a *App) GetLicense() string { return licenseText }

// GetThirdPartyNotices returns the notices of the third-party components.
func (a *App) GetThirdPartyNotices() string { return noticesText }
