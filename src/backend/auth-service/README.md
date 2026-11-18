## Steps on how to generate public and private keys
These keys are necessary for auth-service to work properly.

1. First step is to create directory in root folder:  
```mkdir -p secrets```
  
2. Then generate private key:  
```openssl ecparam -name prime256v1 -genkey -noout -out secrets/jwt_private.pem```
  
3. and public key:  
```openssl ec -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem```   

\
That's it! Congrats, now you can use auth-service on your machine.  
P.S. check if .env file is configured right.