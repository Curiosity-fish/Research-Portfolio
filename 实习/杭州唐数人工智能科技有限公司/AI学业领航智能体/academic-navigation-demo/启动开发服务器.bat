@echo off
chcp 65001 >nul
title 学业领航 - 开发服务器
color 0A

echo ================================================
echo   浙江师范大学 学业领航系统
echo   开发服务器启动工具
echo ================================================
echo.
echo [信息] 正在启动 Vite 开发服务器...
echo.

cd /d "%~dp0"
npm run dev

pause
