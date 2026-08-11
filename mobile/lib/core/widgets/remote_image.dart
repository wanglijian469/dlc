import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../config/app_config.dart';

String absoluteMediaUrl(String? value) {
  if (value == null || value.isEmpty) return '';
  if (value.startsWith('http://') || value.startsWith('https://')) return value;
  return '${AppConfig.apiBaseUrl}$value';
}

class RemoteImage extends StatelessWidget {
  const RemoteImage(this.url,
      {this.height, this.width, this.fit = BoxFit.cover, super.key});
  final String? url;
  final double? height;
  final double? width;
  final BoxFit fit;
  @override
  Widget build(BuildContext context) {
    final resolved = absoluteMediaUrl(url);
    if (resolved.isEmpty) {
      return Container(
          height: height,
          width: width,
          color: const Color(0xFFE8EEF5),
          child: const Icon(Icons.agriculture, color: Color(0xFF6F849C)));
    }
    return CachedNetworkImage(
        imageUrl: resolved,
        height: height,
        width: width,
        fit: fit,
        errorWidget: (_, __, ___) => Container(
            height: height,
            width: width,
            color: const Color(0xFFE8EEF5),
            child: const Icon(Icons.broken_image_outlined)));
  }
}
