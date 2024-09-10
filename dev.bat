@echo off
setlocal enabledelayedexpansion

:: Set error handling
if "%ERRORLEVEL%" NEQ "0" exit /b %ERRORLEVEL%

:: Remove 'dev' directory if it exists, then create it
rmdir /S /Q dev
mkdir dev

:: Change directory to 'dev'
cd dev

:: Clone the repository
git clone https://github.com/scionproto/scion.git

:: Change to the 'scion' directory
cd scion

:: Checkout the specific commit
git checkout 79e7080b3f9471434c3b2e35026ae43dec794c9d

:: Build each component
set CGO_ENABLED=0

cd scion\cmd\scion
go build

cd ..\..\..\scion-pki\cmd\scion-pki
go build

cd ..\..\..\router\cmd\router
go build

cd ..\..\..\control\cmd\control
go build

cd ..\..\..\daemon\cmd\daemon
go build

cd ..\..\..\dispatcher\cmd\dispatcher
go build

:: Go back to the 'dev' directory
cd ..\..\..

:: Create 'bin' directory if it doesn't exist
if not exist bin mkdir bin

:: Copy the built binaries to the 'bin' directory
copy /Y dev\scion\scion\cmd\scion\scion.exe bin\
copy /Y dev\scion\scion-pki\cmd\scion-pki\scion-pki.exe bin\
copy /Y dev\scion\router\cmd\router\router.exe bin\
copy /Y dev\scion\control\cmd\control\control.exe bin\
copy /Y dev\scion\daemon\cmd\daemon\daemon.exe bin\
copy /Y dev\scion\dispatcher\cmd\dispatcher\dispatcher.exe bin\

:: Create 'integration/bin' directory if it doesn't exist
if not exist integration\bin mkdir integration\bin

:: Copy the built binaries to the 'integration/bin' directory
copy /Y dev\scion\scion\cmd\scion\scion.exe integration\bin\
copy /Y dev\scion\scion-pki\cmd\scion-pki\scion-pki.exe integration\bin\
copy /Y dev\scion\router\cmd\router\router.exe integration\bin\
copy /Y dev\scion\control\cmd\control\control.exe integration\bin\
copy /Y dev\scion\daemon\cmd\daemon\daemon.exe integration\bin\
copy /Y dev\scion\dispatcher\cmd\dispatcher\dispatcher.exe integration\bin\

echo Build completed successfully
