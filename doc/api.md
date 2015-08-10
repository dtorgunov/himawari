# Himawari API

Himawari utilises a simple server-client model. Since the main purpose of Himawari is transmitting files over the network, the server will accept requests to transfer (transmitted as simple HTTP messages, with a JSON entity-body describing the file to be sent) and return a URL that will accept said request. If the transmission isn't initiated after a certain amount of time, the URL becomes invalid, and the request needs to be resubmitted.

The connections are to be made through HTTPS. Signing the messages with a TLS or PGP key will serve as the primary means of authentication.

## Transfer request

The client is to connect to the server and transmit a POST request, containing the meta-data of the file about to be transferred, using JSON.

```
{
  "filename": "filename.png",
  "mime": "image/png",
  "length": 12345
}
```

The `filename` field specifies the name the file should have on the server. It is optional, and a randomised name is to be assigned if it is omitted. If a file with that name already exists or a transfer with that filename is already in progress, a 409 message is returned. In the latter case, a header specifies the time after which the current transfer is expected to time out, if it's not in progress already.

The `mime` field specifies the file MIME-type. It is used to assign an extension to a randomised name, if one needs to be generated. The server is allowed to reject requests based on MIME-types with a 403 response.

The `length` field specifies the size of the file about to be transmitted. If it is omitted, a 400 response is issued. The server is allowed to reject requests based on length, if is there is not enough disk space, with a 503 response.

If the request is granted, the server responds with a URL that will accept the file transfer, and return a JSON-encoded reponse with a 200 response code.

```
{
  "url": "http://localhost/filename.png",
  "timeout": 1234,
  "filename": "filename.png"
}
```

The `url` field specifies the address and port that it will accept the transfer on. A PUT request is to be issued to that URL, with the file corresponding to a `filename` request as the entity-body.

If the request is not made within `timeout`, the URL might become invalid.

If the size of the data being transmitted doesn't match the `length` value supplied with the original request, the file may not be created on the server, and a 403 message is returned. This is true if the file is shorter than expected, as well.

If the transfer is completed successfully, a 201 response is sent, with the entity-body and the `Location` header both specifying the URL at which the uploaded file can be accessed.

## Authentication

All messages transmitted to the server need to be cryptographically signed. If the public key corresponding to the key used to sign the request is among the ones authorised by the server, the request is accepted. Otherwise, a 401 message is returned.

The details of how signitures are produced and verifies are to be added at a later point.
