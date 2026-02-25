# Using GitPLM on Windows

(the following was tested on Windows 10. Feel free to open a PR if you have
better ideas or have tested on other versions.)

The GitPLM binaries are not signed, so some effort is required to get past all
the Windows security checks. If someone would like to contribute a process to
sign the binaries for Windows that is not too arduous, that would be
appreciated.

1. Download the appropriate Windows
   [release](https://github.com/git-plm/gitplm/releases). For most people, this
   will be the `windows-x86_64`. If you have a newer ARM based (such as the new
   Surface), you may need the `windows-arm64` release.
1. Extract the downloaded zip file
1. Windows may complain about the program because it is from an Unknown
   publisher. It appears that programs
   [written in Go are often flagged](https://github.com/microsoft/go/issues/1255).
   If Windows pops up a Virus & threat protection dialog, click on the "Severe"
   text, and select "Allow on device", then "Start actions".
1. Double click on the `gitplm` binary in Windows Explorer. Windows may show a
   Windows protected your PC dialog. Click on the "More info" link and select
   "Run Anyway". gitplm should now ask you to enter a directory containing
   partmaster csv files.
1. cp the `gitplm` binary to the `C:\bin` directory.
1. Add `C:\bin` to your system path (System properties->Environment variables)
1. Now in powershell, you should be able to run `gitplm`.
