# Plan: Part Detail Popup and Datasheet Opening

- **Status:** Implemented
- **Starting hash:** `338e9ea`

## What Was Implemented

1. **Open datasheet URL with `o` key.** Works in both normal table view and
   detail popup. Looks up the "Datasheet" column (case-insensitive) and opens
   URLs starting with `http://` or `https://` in the default browser. Shows an
   error message if no valid URL is found.

2. **Platform-aware `openURL` helper** using `xdg-open` (Linux), `open`
   (macOS), or `cmd /c start` (Windows).

3. **Updated help text** in the status bar and detail overlay to show the `o`
   key binding.

## Technical Decisions

- Used `exec.Command().Start()` (non-blocking) so the TUI isn't frozen while
  the browser opens.
- Case-insensitive header match (`strings.EqualFold`) for robustness across
  different CSV files.

## Deviations from Original Plan

- Original plan only had `o` in the detail popup. Extended to also work in
  normal table view so users can open datasheets without entering the detail
  popup first.
