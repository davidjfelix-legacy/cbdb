module Dashboard where


import Html exposing (..)
import Html.Attributes exposing (..)
import Maybe
import Signal exposing (Signal, Address)
import String
import Circuit exposing (Circuit)


type alias Dashboard =
  { circuits : List Circuit }


emptyDashboard : Dashboard
emptyDashboard =
  { circuits = [] }

type Action
  = NoOp
  | AddCircuit Circuit


update : Action -> Dashboard -> Dashboard
update action model =
  case action of
    NoOp -> model
    AddCircuit circuit ->
      { model |
        circuits =
          if String.isEmpty str
            then model.circuits
            else model.circuits ++ [circuit]
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
