@echo off
set PHOTOFOLDER=%HOMEPATH%/Pictures
set LISTEN=:80

set Arch=amd64
if "%PROCESSOR_ARCHITECTURE%" == "x86" ( 
    if not defined PROCESSOR_ARCHITEW6432 set Arch=386
)
set binary=photoboxrecents-windows-%Arch%.exe

%binary% -listen %LISTEN% -photo-folder %PHOTOFOLDER%