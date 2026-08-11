import 'dart:io';

import 'package:yaml/yaml.dart';

void main(List<String> arguments) {
  if (arguments.length != 1) {
    stderr.writeln('Usage: dart run validate_openapi.dart <openapi.yaml>');
    exitCode = 64;
    return;
  }

  final contract = File(arguments.single);
  if (!contract.existsSync()) {
    stderr.writeln('OpenAPI contract not found: ${contract.path}');
    exitCode = 66;
    return;
  }

  final document = loadYaml(contract.readAsStringSync());
  if (document is! YamlMap || document['openapi'] == null) {
    stderr.writeln('Contract must be an OpenAPI YAML document.');
    exitCode = 65;
    return;
  }

  final paths = document['paths'];
  final schemas = (document['components'] as YamlMap?)?['schemas'];
  if (paths is! YamlMap || paths.isEmpty || schemas is! YamlMap || schemas.isEmpty) {
    stderr.writeln('Contract must define non-empty paths and component schemas.');
    exitCode = 65;
    return;
  }

  const requiredPaths = <String>{
    '/api/v1/app/auth/login',
    '/api/v1/app/auth/refresh',
    '/api/v1/market-posts',
    '/api/v1/market-posts/{id}/contact',
    '/api/v1/me/market-posts',
    '/api/v1/me/profile',
    '/api/v1/media',
    '/api/admin/market-posts',
    '/api/admin/market-posts/{id}/status',
  };
  final missing = requiredPaths.where((path) => !paths.containsKey(path)).toList();
  if (missing.isNotEmpty) {
    stderr.writeln('Contract is missing required paths: ${missing.join(', ')}');
    exitCode = 65;
    return;
  }

  stdout.writeln(
      'Validated OpenAPI ${document['openapi']}: ${paths.length} paths, ${schemas.length} schemas.');
}
