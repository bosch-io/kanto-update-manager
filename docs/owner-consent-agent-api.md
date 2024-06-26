### Overview
The Update Manager interacts with a Human Machine Interface(HMI) application using MQTT messages with JSON payload. These messages implement the designs described in [Owner Consent Specification](./owner-consent-specification.md). 

### Message Format
The messages for the bidirectional exchange between the Update Manager and the HMI application are carried in the following format:

```
{
  "activityId": "123e4567-e89b-12d3-a456-426614174000",
  "timestamp": 123456789,
  "payload": {} // actual message content as per message specification
}
```

### Message Data Model
The message data model has the following three metadata elements:

- `activityId` [string]: UUID generated by the backend which is used for correlating a set of device / backend messages with an activity entity (e.g. a desired state application process) on system level

- `timestamp` [int64]: Message creation timestamp. Number of milliseconds that have elapsed since the Unix epoch (00:00:00 UTC on 1 January 1970)

- `payload` [object]: Custom, unstructured message payload per message specification

### MQTT Topics
The Update Manager and the connector component bidirectionally exchange messages in the previously described format using the following MQTT topics:

| Topic | Direction | Purpose |
| - | - | - |
| `${some-optional-prefix}update/ownerconsent` | Update Manager -> HMI | Informing the HMI application that an owner consent is needed to proceed with the update process |
| `${some-optional-prefix}update/ownerconsentfeedback` | HMI -> Update Manager | Informing the Update Manager that the owner approved / denied the update |

`${some-optional-prefix}` can be any string defined for the concrete deployment, e.g. `device`, `vehicle`, etc.
