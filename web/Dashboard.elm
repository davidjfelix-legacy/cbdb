module Dashboard where


import Html exposing (..)
import Html.Attributes exposing (..)
import Maybe
import Signal exposing (Signal, Address)
import String


type alias Dashboard =
  { circuits : List Circuit }


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


emptyDashboard : Dashboard
emptyDashboard =
  { circuits = [] }

type Action
  = NoOp
  | AddCircuit String


update : Action -> Dashboard -> Dashboard
update action model =
  case action of
    NoOp -> model
    AddCircuit str ->
      { model |
        circuits =
          if String.isEmpty str
            then model.circuits
            else model.circuits ++ [str]
      }



view : Address Action -> Dashboard -> Html
view action model =
  div
    [ class "dashboard-wrapper"]
    [ text "derp" ]



main : Signal Html
main =
  Signal.map (view actions.address) model


model : Signal Dashboard
model =
  Signal.foldp update initialModel actions.signal


initialModel : Dashboard
initialModel =
  emptyDashboard


actions : Signal.Mailbox Action
actions =
  Signal.mailbox NoOp
