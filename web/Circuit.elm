module Circuit where

import Html exposing (..)
import Html.Attributes exposing (..)
import Maybe
import Signal exposing (Signal, Address)
import String

type alias Circuit =
  { serviceName: String
  , circuitName: String
  , currentStatus: BreakerStatus
  , currentLatency: LatencyHistogram
  , currentCount: CircuitCounter
  , historicInformation: List CircuitHistory
  }

type alias CircuitHistory =
  { latency: LatencyHistogram
  , count: CircuitCounter
  }


type alias CircuitCounter =
  { total: Int
  , failures: Maybe Int
  , success: Maybe Int
  , shortCircuited: Maybe Int
  , failover: Maybe Int
  }


type alias LatencyHistogram =
  { mean: Int
  , median: Int
  , min: Int
  , max: Int
  , percentile25: Maybe Int
  , percentile75: Maybe Int
  , percentile90: Maybe Int
  , percentile95: Maybe Int
  , percentile99: Maybe Int
  , percentile995: Maybe Int
  , percentile999: Maybe Int
  }


type alias PartiallyOpenBreaker =
  { open: Int
  , closed: Int
  }


type BreakerStatus
  = Open Int
  | Partial PartiallyOpenBreaker
  | Closed Int


type Action
  = NoOp

update: Action -> Circuit -> Circuit
update action model =
  case action of
    Noop -> model
    
