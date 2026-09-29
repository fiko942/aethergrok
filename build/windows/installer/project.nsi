Unicode true

####
## AetherGrok Desktop Studio - Windows NSIS Installer Script
####
## The following information is taken from ProjectInfo / defaults, but can be overwritten here.
####
!ifndef INFO_PROJECTNAME
    !define INFO_PROJECTNAME    "aethergrok"
!endif
!ifndef INFO_COMPANYNAME
    !define INFO_COMPANYNAME    "AetherGrok"
!endif
!ifndef INFO_PRODUCTNAME
    !define INFO_PRODUCTNAME    "AetherGrok"
!endif
!ifndef INFO_PRODUCTVERSION
    !define INFO_PRODUCTVERSION "1.0.5"
!endif
!ifndef INFO_COPYRIGHT
    !define INFO_COPYRIGHT      "Copyright (c) 2026 Wiji Fiko Teren"
!endif
!ifndef PRODUCT_EXECUTABLE
    !define PRODUCT_EXECUTABLE  "${INFO_PROJECTNAME}.exe"
!endif
!ifndef UNINST_KEY_NAME
    !define UNINST_KEY_NAME     "${INFO_PRODUCTNAME}"
!endif
!ifndef REQUEST_EXECUTION_LEVEL
    !define REQUEST_EXECUTION_LEVEL "admin"
!endif

####
## Include the wails tools
####
!include "wails_tools.nsh"

# Version information resource
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"      "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription"  "${INFO_PRODUCTNAME} Setup"
VIAddVersionKey "ProductVersion"   "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"      "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"   "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"      "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_UNFINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

# Finish page options - run application checkbox
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Launch ${INFO_PRODUCTNAME}"

# Installer Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

# Uninstaller Pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"

!ifndef OUTFILE_NAME
    !define OUTFILE_NAME "..\..\bin\AetherGrok-Setup.exe"
!endif
OutFile "${OUTFILE_NAME}"

!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
    InstallDirRegKey HKCU "${UNINST_KEY}" "InstallLocation"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
    InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
  InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"
!endif

ShowInstDetails show
ShowUninstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section "MainSection" SEC01
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$LOCALAPPDATA\${INFO_PROJECTNAME}"
    RMDir /r "$APPDATA\${INFO_PROJECTNAME}"

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller

    RMDir /r "$INSTDIR"
SectionEnd
