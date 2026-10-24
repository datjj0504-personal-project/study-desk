#ifndef AppVersion
  #error AppVersion must be supplied by release-windows.ps1
#endif
#ifndef AppExePath
  #error AppExePath must be supplied by release-windows.ps1
#endif
#ifndef WebView2BootstrapperPath
  #error WebView2BootstrapperPath must be supplied by release-windows.ps1
#endif
#ifndef OutputDirectory
  #error OutputDirectory must be supplied by release-windows.ps1
#endif

[Setup]
AppId={{7A6BAF72-4CB6-4A44-9D3A-C0C18A0C7A62}
AppName=StudyDesk
AppVersion={#AppVersion}
AppVerName=StudyDesk {#AppVersion}
AppPublisher=StudyDesk
DefaultDirName={autopf}\StudyDesk
DefaultGroupName=StudyDesk
UninstallDisplayName=StudyDesk
UninstallDisplayIcon={app}\studydesk.exe
MinVersion=10.0
OutputDir={#OutputDirectory}
OutputBaseFilename=StudyDesk-Setup-{#AppVersion}
VersionInfoVersion={#AppVersion}
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
DisableProgramGroupPage=yes
WizardStyle=modern
Compression=lzma2
SolidCompression=yes

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "{#AppExePath}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#WebView2BootstrapperPath}"; DestDir: "{tmp}"; DestName: "MicrosoftEdgeWebView2Setup.exe"; Flags: deleteafterinstall

[Icons]
Name: "{autoprograms}\StudyDesk"; Filename: "{app}\studydesk.exe"
Name: "{autodesktop}\StudyDesk"; Filename: "{app}\studydesk.exe"; Tasks: desktopicon

[Run]
Filename: "{tmp}\MicrosoftEdgeWebView2Setup.exe"; Parameters: "/silent /install"; StatusMsg: "Installing or updating Microsoft Edge WebView2 Runtime..."; Flags: waituntilterminated runhidden
