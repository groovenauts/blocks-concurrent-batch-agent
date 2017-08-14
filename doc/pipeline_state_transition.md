# Pipeline state transition

## State Transition Diagram

![Pipeline state transition](./pipeline_state_transition.png)

## State Transition Table

| No. | Request path                 | Method             | (Start) | Uninitialized | Broken | Pending  | Reserved | Building  | Deploying | Opened   | Closing | ClosingError | Closed |
|----:|-----------------------------|---------------------|:-------:|:-------------:|:------:|:--------:|:--------:|:---------:|:---------:|:-------:|:--------:|:------------:|:------:|
|  1  | /pipelines/                 | ReserveOrWait       | Reserved/Pending | -    |     -  |    -     |     -    |    -      |    -      |    -     |    -    |     -        | -      |
|  2  | /pipelines/:id/refresh_task | (another instance).CompleteClosing | - | N/A  | N/A    | Reserved | N/A      | N/A       | N/A       | N/A      | N/A     | N/A          | N/A    |
|  3  | /pipelines/:id/build_task   | StartBuilding       | -       | N/A           | N/A    | N/A      | Building | Building  | N/A       | N/A     | N/A      | N/A          | N/A    |
|  4  | /pipelines/:id/build_task   | StartDeploying      | -       | N/A           | N/A    | N/A      | N/A      | Deploying | N/A       | N/A     | N/A      | N/A          | N/A    |
|  5  | /pipelines/:id/refresh_task | FailDeploying       | -       | N/A           | N/A    | N/A      | N/A      | N/A       | Broken    | N/A     | N/A      | N/A          | N/A    |
|  6  | /pipelines/:id/refresh_task | CompleteDeploying   | -       | N/A           | N/A    | N/A      | N/A      | N/A       | Opened    | N/A     | N/A      | N/A          | N/A    |
|  7  | /pipelines/:id/close_task   | StartClosing        | -       | N/A           | N/A    | N/A      | N/A      | N/A       | N/A       | Closing | Closing      | N/A      | N/A    |
|  8  | /pipelines/:id/refresh_task | FailClosing         | -       | N/A           | N/A    | N/A      | N/A      | N/A       | N/A       | N/A     | ClosingError | N/A      | N/A    |
|  9  | /pipelines/:id/refresh_task | CompleteClosing     | -       | N/A           | N/A    | N/A      | N/A      | N/A       | N/A       | N/A     | Closed       | N/A      | N/A    |
