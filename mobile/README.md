# 大陆农机配件 Flutter App

Android-first Flutter client for the shared Gin REST API. The application uses
Riverpod, GoRouter, Dio and platform secure storage. It never connects to MySQL
directly and never logs passwords, tokens or contact phone numbers.

## Bootstrap

The repository pins Flutter 3.44.9 in `.fvmrc` and includes Android and iOS
platform projects. Install that Flutter version and an Android SDK, then run:

```powershell
flutter pub get
```

Development against an Android emulator:

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

Production builds must set an HTTPS endpoint:

```powershell
flutter build apk --release --dart-define=PRODUCTION=true --dart-define=API_BASE_URL=https://example.com
```

Run `flutter analyze`, `flutter test`, and `flutter build apk --debug` before
delivery. iOS is kept build-compatible but signing and App Store delivery are
outside the first release.
