# Pipeline state transition

## State Transition Diagram

![Pipeline state transition](./pipeline_state_transition.png)

## State Transition Table

| No. | Request path                     |  (Start) | Uninitialized | Broken | Pending  | Waiting  | Reserved | Building  | Deploying | Opened   | 
|----:|----------------------------------|:--------:|:-------------:|:------:|:--------:|:--------:|:--------:|:---------:|:---------:|:--------:|
|  1  | POST /pipelines/                 | Pending/Waiting/Reserved | -      |  - | -   |     -    |     -    |    -      |    -      |    -     |
|  2  | (Dependency DONE)                |  -       | N/A           | N/A    | Waiting/Reserved | N/A | N/A   | N/A       | N/A       | N/A      | 
|  3  | (Dependency DONE)                |  -       | N/A           | N/A    | N/A      | Reserved | N/A      | N/A       | N/A       | N/A      | 
|  4  | POST /pipelines/:id/build_task   |  -       | N/A           | N/A    | N/A      | N/A      | Building->Deploying  | N/A | N/A | N/A      | 
|  5  | POST /pipelines/:id/wait_building_task |  - | N/A           | N/A    | N/A      | N/A      | N/A      | N/A       | Opened    | N/A      | 

| No. | Request path                                                  | Opened   | Closing  | ClosingError | Closed |
|----:|---------------------------------------------------------------|:--------:|:--------:|:------------:|:------:|
|  6  | POST /pipelines/:id/subscribe_task (NOT DONE)                 | -        | N/A      | N/A          | N/A    |
|  7  | POST /pipelines/:id/subscribe_task (DONE with hibernation)    | HibernationChecking | N/A | N/A    | N/A    |
|  8  | POST /pipelines/:id/subscribe_task (DONE without hibernation) | Closing  | N/A      | N/A          | N/A    |
|  9  | POST /pipelines/:id/wait_closing_task (NOT DONE)              | N/A      | -        | N/A          | N/A    |
| 10  | POST /pipelines/:id/wait_closing_task (ERROR)                 | N/A      | ClosingError  | N/A     | N/A    |
| 11  | POST /pipelines/:id/wait_closing_task (DONE)                  | N/A      | Closed   | N/A          | N/A    |

| No. | Request path                                                  | HibernationChecking   | HibernationStarting   | HibernationProcessing | HibernationError | Hibernating |
|----:|---------------------------------------------------------------|:---------------------:|:---------------------:|:---------------------:|:----------------:|:-----------:|
| 12  | POST /pipelines/:id/check_hibernation_task                    | HibernationStarting   | N/A                   | N/A                   | N/A               | N/A        |
| 13  | POST /pipelines/:id/hibernate_task                            | N/A                   | HibernationProcessing | N/A                   | N/A               | N/A        |
| 14  | POST /pipelines/:id/wait_hibernation_task (NOT DONE)          | N/A                   | | N/A                 | -                     | N/A               | N/A        |
| 15  | POST /pipelines/:id/wait_hibernation_task (ERROR)             | N/A                   | | N/A                 | HibernationError      | N/A               | N/A        |
| 16  | POST /pipelines/:id/wait_hibernation_task (DONE)              | N/A                   | | N/A                 | Hibernating           | N/A               | N/A        |
| 17  | POST /pipelines/:id/jobs                                      | N/A                   | | N/A                 | N/A                   | N/A               | Building   |
