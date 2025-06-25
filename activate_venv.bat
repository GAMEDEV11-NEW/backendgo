@echo off
echo Activating Redis Socket.IO Virtual Environment...
echo ===============================================

call venv\Scripts\activate.bat

echo.
echo ✅ Virtual environment activated!
echo.
echo 🐍 Python: %VIRTUAL_ENV%\Scripts\python.exe
echo 📦 You can now run:
echo    - python quick_test.py
echo    - python test_trigger_flow.py
echo    - python game_list_updater.py
echo.
echo 💡 To deactivate, type: deactivate
echo.

cmd /k 