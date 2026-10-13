# Name the orchestration package `lifecycle`

The package that starts and stops providers is `lifecycle` (`lifecycle.Run` for Wire composition, `lifecycle.Group` for manual composition), replacing the original `stack` and the draft `container`. `container` reads like a DI container sitting next to Wire, `app` collides with the conventional `App`, `NewApp` and `InitApp` names of Wire applications, and `kernel` suggests request dispatch rather than start-up and shutdown.
