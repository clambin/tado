/*
Package tado provides an API Client for the Tadoº smart thermostat devices

NOTE: the tado package currently only supports heating devices.  Hot water & AC devices are not supported.
If you have access to these devices, let me know, so I can add support for these in a later release.

# Multi-home accounts

Most Tado users will only have a single home associated with their Tado account. If an account has multiple homes,
this package will by default use the first home associated with the account. To use the package's API for another home,
use SetActiveHomeByName to set the active home. All subsequent commands will be executed against that home.
*/
package tado
