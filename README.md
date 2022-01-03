# Parsley Homework (Backend)
**Author: Brandon Raphael**

## Utilizing the Application
run the application with: `make run`

## Considerations & Assumptions

* How will the application handle times which typically would not be schedulable? (e.g. holidays, weekends, night time, lunch time, PTO, etc.)
  * First approach does not handle these. They could be added as conditionals, however, to the isAvailable method, as this gets used by the scheduler.


* Is "availability" representative of everyone in the system, or should we have individual availability as a feature? (e.g. should we be able to schedule for different doctors, nurses, etc.)
  * First approach does not implement individual scheduling (however, I have included attendants and requestors as fields in all of the API endpoints for future development.)


* How will the application handle Time Zone differences? (e.g. the customer wants to schedule an appointment for 3pm PST, with their doctor being located in an EST time zone)
  * First approach would be to convert everything to Universal time.


* Which time intervals should be schedulable? And how long are appointments? (e.g. the beginning of every hour, every thirty minutes, every fifteen minutes? etc.)
  * Assumption made here will be that as long as the start and end times do not contain and other appointments, they should be schedulable (down to the second).


* How do we define duration on input?
  * Users giving a start time and a duration seems counterintuitive. Whereas durations can have multiple definitions, in this context it might be easiest to have the user request a start and end time/date of the same formatting.


* What is the maximum and minimum duration of an appointment?
  * Assumption made here is that maximum is no more than say 8 hours (one business day). Minimum is perhaps 15 minutes.


* How do we best (most efficiently and with which data structure) model the schedule?
  * I've created a struct to contain start and end times, as well as requestor and attendant for schedules. These structs are stored, unordered, as lists which are the values in a map. The map's keys are day-date strings.


* What endpoints are needed and what will they look like?
  * Defined in the prompt essentially. Reserve, release, and check availability. They are defined in more detail below.


* What will the inputs look like for each of the endpoints? If there exists a format for these inputs, what will they be?
  * The inputs are the same for each endpoint (for simplicity's sake here, I might instead use a GET for the availability in a real world situation). They are defined further below.


## API
### :8080/availability (POST)
Retrieves timeslot availability.

Sample JSON input body:
```
{
    "start":"2014-09-12T11:45:26.371Z",
    "end":"2014-09-12T12:45:26.371Z",
    "requestor":"me",
    "attendant":"you"
}
```

Sample output:
```
200
false
```

### :8080/reserve (POST)
Reserves a timeslot. Returns 400 if timeslot was already taken.

Sample JSON input body:
```
{
    "start":"2014-09-12T11:45:26.371Z",
    "end":"2014-09-12T12:45:26.371Z",
    "requestor":"me",
    "attendant":"you"
}
```

Sample output:
```
200
Reservation confirmed.
```

### :8080/release (POST)
Releases an appointment. Returns 400 if timeslot could not be found.

Sample JSON input body:
```
{
    "start":"2014-09-12T11:45:26.371Z",
    "end":"2014-09-12T12:45:26.371Z",
    "requestor":"me",
    "attendant":"you"
}
```

Sample output:
```
200
Release confirmed.
```

### :8080/schedule (GET)
Gets the current schedule map.

Sample output:
```
200
"map[2014-09-12:[{2014-09-12 13:45:26.371 +0000 UTC 2014-09-12 14:46:26.371 +0000 UTC me you}]]"
```
