# Library code logs through log/slog

Every package logs through the standard `log/slog` default logger. The `log/zap` and `log/logrus` providers route slog into their backend on Init and restore the previous logger on Close, so a service gets one log stream whichever backend it picks, and the core packages no longer depend on logrus.
