## Steps on how to generate public and private keys
These keys are necessary for auth-service to work properly.

> Prerequisite (Windows only): Make sure OpenSSL is installed and on your PATH.

1. Create the secrets directory in src/backend   
    Linux / macOS:   
    ```mkdir -p secrets```

    Windows (CMD):   
    ```mkdir secrets```

    Windows (PowerShell):   
    ```New-Item -ItemType Directory -Path secrets -Force```

2. Generate private key   
```openssl ecparam -name prime256v1 -genkey -noout -out secrets/jwt_private.pem```
   
3. Generate public key   
```openssl ec -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem```
   

\
That's it! Congrats, now you can use auth-service on your machine.

> P.S. Check if .env file is configured right.