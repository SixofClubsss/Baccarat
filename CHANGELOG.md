# Changelog

This file lists the changes to Baccarat repo with each version.

## 0.x.x - In Progress

### Added 
* Hand log
* FetchLastHand(), show last hand when connected


### Changed
* Rename FetchBaccHand() to FetchHand
* Handle Action Show/Hide in OnTapped 
* Clean up unused vars


## 0.3.1 - January 19 2024

### Changed
* Go 1.21.5
* Fyne 2.4.3
* dReams 0.11.1
* Cleaned up `rpc` client var names


## 0.3.0 - December 23 2023

### Added

* CHANGELOG
* Pull request and bug templates
* `semver` versioning 
* Stand alone dApp theme
* Asset tab with profile

### Changed

* Fyne 2.4.1
* dReams 0.11.0
* Icon resources 
* Button UI updates
* Remove swap and import from `holdero`
* implement `gnomes` and funcs
* implement `rpc` PrintError, PrintLog and IsConfirmingTx

### Fixed

* Deprecated container.NewMax
* Fyne error when downloading custom cards
* Validator hang