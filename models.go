package main

import (
  "time"
)

type TimeSlotReq struct {
  Start string
  End string
  Requestor string
  Attendant string
}

type TimeSlot struct {
  Start time.Time
  End time.Time
  Requestor string
  Attendant string
}
