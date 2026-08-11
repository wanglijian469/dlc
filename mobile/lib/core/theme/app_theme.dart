import 'package:flutter/material.dart';

class AppTheme {
  const AppTheme._();
  static const navy = Color(0xFF123F73);
  static const orange = Color(0xFFD87A2B);
  static const surface = Color(0xFFF3F6FA);

  static ThemeData get light => ThemeData(
        colorScheme: ColorScheme.fromSeed(
            seedColor: navy,
            primary: navy,
            secondary: orange,
            surface: surface),
        scaffoldBackgroundColor: surface,
        useMaterial3: true,
        appBarTheme: const AppBarTheme(
            backgroundColor: Colors.white,
            foregroundColor: Color(0xFF172B43),
            elevation: 0),
        cardTheme: const CardThemeData(
            color: Colors.white, elevation: 0, margin: EdgeInsets.zero),
        inputDecorationTheme: InputDecorationTheme(
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(color: Color(0xFFD8E2EE))),
            enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(color: Color(0xFFD8E2EE)))),
      );
}
