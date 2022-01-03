package main

import (
  "fmt"
  "time"
  "log"
  "github.com/gorilla/mux"
  "net/http"
  "encoding/json"
)

var Schedule = make(map[string][]TimeSlot)

func main() {
  fmt.Println("Running scheduling application...")

  r := mux.NewRouter()
  r.HandleFunc("/availability", handleAvailabilityRequest).Methods("POST")
  r.HandleFunc("/reserve", handleReserveRequest).Methods("POST")
  r.HandleFunc("/release", handleReleaseRequest).Methods("POST")
  r.HandleFunc("/schedule", handleScheduleRequest).Methods("GET")

  log.Fatal(http.ListenAndServe(":8080", r))
}

func validateTimes(t TimeSlot) (error) {
  if t.Start.After(t.End) {
    return fmt.Errorf("TimeSlot START came after TimeSlot END")
  }

  difference := t.End.Sub(t.Start)
  if difference.Hours() > 7 && difference.Minutes() > 0 {
    return fmt.Errorf("TimeSlot duration exceeded maximum allowable duration of appointment (8 hours).")
  }

  if difference.Hours() == 0 && difference.Minutes() < 15 {
    return fmt.Errorf("TimeSlot duration was less than minimum allowable duration of appointment (15 minutes).")
  }

  return nil
}

func isAvailable(t TimeSlot) (bool) {
  startDate := t.Start.Format("2006-01-02")
  endDate :=  t.End.Format("2006-01-02")

  // If the date string already exists in the map, check there are no pre-existing
  // overlapping appointments on that day
  if slots, ok := Schedule[startDate]; ok {
    for _, slot := range slots {
      if t.Start.Before(slot.Start) && t.Start.Before(slot.End) {
        if t.End.After(slot.Start) && t.End.Before(slot.End) {
          return false
        }
        if t.End.After(slot.Start) && t.End.After(slot.End) {
          return false
        }
      } else if t.Start.After(slot.Start) && t.Start.Before(slot.End) {
        if t.End.After(slot.Start) && t.End.Before(slot.End) {
          return false
        }
        if t.End.After(slot.Start) && t.End.After(slot.End) {
          return false
        }
      } else if t.Start.Equal(slot.Start) || t.End.Equal(slot.End) {
        return false
      }
    }
  }

  // If the end of the appointment is on the next day (UTC) we need to check if
  // there exists data on the next day already, and if so, if there are any pre-
  // existing overlapping appointments on that day.
  if startDate != endDate {
    if slots, ok := Schedule[endDate]; ok {
      for _, slot := range slots {
        if t.Start.Before(slot.Start) && t.Start.Before(slot.End) {
          if t.End.After(slot.Start) && t.End.Before(slot.End) {
            return false
          }
          if t.End.After(slot.Start) && t.End.After(slot.End) {
            return false
          }
        }
      }
    }
  }

  // Finally, we need to check if there were any appointments for the day prior
  // which could have endtimes today the might overlap with the new appointment.
  // This would be a difficult check to make if I hadn't required appointment start
  // and end times to be within 8 hours of one another.
  yesterday := t.Start.AddDate(0, 0, -1).Format("2006-01-02")
  if slots, ok := Schedule[yesterday]; ok {
    for _, slot := range slots {
      if t.Start.After(slot.Start) && t.Start.Before(slot.End) {
        return false
      }
    }
  }

  return true
}

func reserveAvailability(t TimeSlot) string {
  startDate := t.Start.Format("2006-01-02")
  if _, ok := Schedule[startDate]; ok {
    Schedule[startDate] = append(Schedule[startDate], t)
  } else {
    Schedule[startDate] = []TimeSlot{t}
  }

  return "Reservation confirmed."
}

func getTimeSlot(w http.ResponseWriter, r *http.Request) (TimeSlot, error) {
  timeSlot := TimeSlot {}
  w.Header().Set("Content-Type", "application/json")
  var timeSlotReq TimeSlotReq
  json.NewDecoder(r.Body).Decode(&timeSlotReq)

  layout := time.RFC3339
  start, err := time.Parse(layout, timeSlotReq.Start)
  if err != nil {
    fmt.Println(err)
    return timeSlot, fmt.Errorf("Requested START time does not match format. Must follow formatting, in UTC time: 'YYYY-MM-DDTHH:MM:SS.SSSZ' e.g. '2014-09-12T11:45:26.371Z'. Error is: `%v`", err)
  }

  end, err := time.Parse(layout, timeSlotReq.End)
  if err != nil {
    return timeSlot, fmt.Errorf("Requested END time does not match format. Must follow formatting, in UTC time: 'YYYY-MM-DDTHH:MM:SS.SSSZ' e.g. '2014-09-12T11:45:26.371Z'. Error is: `%v`", err)
  }

  timeSlot = TimeSlot {
    Start: start,
    End: end,
    Requestor: timeSlotReq.Requestor,
    Attendant: timeSlotReq.Attendant,
  }

  err = validateTimes(timeSlot)
  if err != nil {
    return timeSlot, fmt.Errorf("Error validating start and end times, error is: `%v`", err)
  }

  return timeSlot, nil
}

func removeIndex(s []TimeSlot, index int) []TimeSlot {
    return append(s[:index], s[index+1:]...)
}

func handleAvailabilityRequest(w http.ResponseWriter, r *http.Request) {
  timeSlot, errRes := getTimeSlot(w, r)
  if errRes != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    json.NewEncoder(w).Encode(errRes)
    return
  }

  if isAvailable(timeSlot) {
    json.NewEncoder(w).Encode(true)
  } else {
    json.NewEncoder(w).Encode(false)
  }
}

func handleReserveRequest(w http.ResponseWriter, r *http.Request) {
  timeSlot, errRes := getTimeSlot(w, r)
  if errRes != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    json.NewEncoder(w).Encode(errRes)
    return
  }

  if isAvailable(timeSlot) {
    json.NewEncoder(w).Encode(reserveAvailability(timeSlot))
  } else {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    json.NewEncoder(w).Encode("Failed to reserve requested timeslot. Timeslot was not available.")
  }
}

func handleReleaseRequest(w http.ResponseWriter, r *http.Request) {
  t, errRes := getTimeSlot(w, r)
  if errRes != nil {
    http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
    json.NewEncoder(w).Encode(errRes)
    return
  }

  startDate := t.Start.Format("2006-01-02")

  if slots, ok := Schedule[startDate]; ok {
    for i, slot := range slots {
      if t.Start.Equal(slot.Start) && t.End.Equal(slot.End) {
        Schedule[startDate] = removeIndex(Schedule[startDate], i)
        json.NewEncoder(w).Encode("Release Confirmed.")
        return
      }
    }
  }

  http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
  json.NewEncoder(w).Encode("Release failed. Appointment could not be found.")
}

func handleScheduleRequest(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(fmt.Sprint(Schedule))
}
