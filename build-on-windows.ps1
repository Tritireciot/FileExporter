$workingDir = "./standalone-windows-64";
$assets = "./assets";
$version = "weasyprint==68.1"

if (Test-Path $workingDir) {
    Write-Host "*** Cleaning $workingDir"
    Remove-Item $workingDir -Recurse -Force | Out-Null
}
Write-Host "*** Creating $workingDir"
New-Item -ItemType "directory" -Path $workingDir | Out-Null

if (!(Test-Path $assets)) {
    Write-Host "*** Creating $assets"
    New-Item -ItemType "directory" -Path $assets | Out-Null
}

$workingDir = Resolve-Path -Path "./standalone-windows-64";
$assets = Resolve-Path -Path "./assets";

Write-Host "*** Downloading msys2 environment"
Invoke-WebRequest -Uri "https://github.com/msys2/msys2-installer/releases/download/2025-12-13/msys2-base-x86_64-20251213.sfx.exe" -OutFile "$workingDir/msys2-base-x86_64-20251213.sfx.exe"
Write-Host "*** Extracting msys2 environment"
Invoke-Expression "$workingDir/msys2-base-x86_64-20251213.sfx.exe -y -o$workingDir"
Remove-Item "$workingDir/msys2-base-x86_64-20251213.sfx.exe" -Recurse -Force | Out-Null

Write-Host "*** Downloading python (https://github.com/indygreg/python-build-standalone)"
Invoke-WebRequest -Uri "https://github.com/astral-sh/python-build-standalone/releases/download/20260408/cpython-3.13.13+20260408-x86_64-pc-windows-msvc-install_only.tar.gz" -OutFile "$workingDir/python.tar.gz"
Write-Host "*** Extracting python"
Invoke-Expression "tar -xvzf $workingDir/python.tar.gz -C $workingDir"
Remove-Item "$workingDir/python.tar.gz" -Recurse -Force | Out-Null

Write-Host "*** Installing weasyprint dependencies"
Invoke-Expression "$workingDir\msys64\usr\bin\bash -lc 'pacman -S mingw-w64-x86_64-pango mingw-w64-x86_64-sed --noconfirm'"

Write-Host "*** Remove python.exe to avoid conflicts with msys2 python"
Remove-Item "$workingDir\msys64\mingw64\bin\python.exe"

Write-Host "*** Installing weasyprint and pyinstaller"
Invoke-Expression "$workingDir\python\python.exe -m pip install weasyprint==68.1 pyinstaller"

Write-Host "*** Patching weasyprint __main__.py to fix imports"
Invoke-Expression "$workingDir\msys64\mingw64\bin\sed -i 's/^from \. /from weasyprint /' $workingDir/python/Lib/site-packages/weasyprint/__main__.py"
Invoke-Expression "$workingDir\msys64\mingw64\bin\sed  -i 's/^from \./from weasyprint\./' $workingDir/python/Lib/site-packages/weasyprint/__main__.py"
$Env:PATH = "$workingDir\msys64\mingw64\bin;$Env:PATH"

Write-Host "*** Building weasyprint executable"
Invoke-Expression "$workingDir\python\python.exe -m PyInstaller $workingDir/python/Lib/site-packages/weasyprint/__main__.py -n weasyprint -D"

Write-Host "*** Cleaning up python and msys2 environments for testing weasyprint executable"
Remove-Item "$workingDir/python" -Recurse -Force | Out-Null
Remove-Item "$workingDir/msys64" -Recurse -Force | Out-Null

Set-Location  "./dist/"
Move-Item -Path "./weasyprint" -Destination "./weasyprint-windows"
New-Item -Path "version-$version"

Set-Location  "./weasyprint-windows"
Write-Host "*** Testing weasyprint"
Invoke-Expression ".\weasyprint.exe --info"

$targetGoPackageDir = "../PrintServer/transformer"

Set-Location "./dist/"

# Переименовываем собранную папку dist/weasyprint в weasyprint-windows
Write-Host "*** Preparing executable folder..."
Move-Item -Path "./weasyprint" -Destination "./weasyprint-windows"

# Краткий тест работоспособности бинарника
Set-Location "./weasyprint-windows"
Write-Host "*** Testing weasyprint executable..."
Invoke-Expression ".\weasyprint.exe --info"
Set-Location "../" # возвращаемся в папку dist

# 2. Очищаем старую сборку в Go-пакете, если она там осталась с прошлого раза
$finalDestination = Join-Path $targetGoPackageDir "weasyprint-windows"
if (Test-Path $finalDestination) {
    Write-Host "   ⚠️ Removing old build from Go package..."
    Remove-Item $finalDestination -Recurse -Force | Out-Null
}

# 3. Мгновенно перемещаем папку weasyprint-windows напрямую в ваш Go-проект
Write-Host "🚚 Moving compiled weasyprint-windows directly to Go package..."
Move-Item -Path "./weasyprint-windows" -Destination $targetGoPackageDir

# Возвращаемся в корень проекта
Set-Location "../"

# 4. Финальная чистка временных артефактов сборщика PyInstaller
Write-Host "🧹 Cleaning up temporary build directories..."
if (Test-Path "./dist") { Remove-Item -Path "./dist" -Recurse -Force | Out-Null }
if (Test-Path "./build") { Remove-Item -Path "./build" -Recurse -Force | Out-Null }
if (Test-Path $workingDir) { Remove-Item -Path $workingDir -Recurse -Force | Out-Null }
if (Test-Path weasyprint.spec) { Remove-Item -Path weasyprint.spec -Recurse -Force | Out-Null }