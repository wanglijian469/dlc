class AppConfig {
  const AppConfig._();

  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080',
  );

  static const logNetwork = bool.fromEnvironment('LOG_NETWORK');

  static void validate() {
    const production = bool.fromEnvironment('PRODUCTION');
    if (production && !apiBaseUrl.startsWith('https://')) {
      throw StateError('Production API_BASE_URL must use HTTPS.');
    }
  }
}
