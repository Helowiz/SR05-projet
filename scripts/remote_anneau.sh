#!/bin/bash

APP=$1
CTL=$2
ROLE=$3
MAC_IP=$4  # Only required when running the WSL role

PORT_C2_TO_C3=61234
PORT_C3_TO_C1=61235

# 1. Validate arguments and define FIFOs based on role
if [ "$ROLE" == "wsl" ]; then
    if [ -z "$MAC_IP" ]; then
        echo "Error: WSL role requires the Mac's IP address."
        echo "Usage: $0 <APP> <CTL> wsl <MAC_IP>"
        exit 1
    fi
    FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
           /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2)

elif [ "$ROLE" == "mac" ]; then
    FIFOS=(/tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3)

else
    echo "Usage: $0 <APP> <CTL> <wsl|mac> [MAC_IP]"
    exit 1
fi

# 2. Cleanup function
cleanup() {
    echo -e "\nCleaning up $ROLE..."
    kill $(jobs -p) 2>/dev/null
    rm -f "${FIFOS[@]}"
}
trap cleanup EXIT INT TERM

# 3. Create FIFOs
mkfifo "${FIFOS[@]}"


# 4. Execute topology based on Role
if [ "$ROLE" == "mac" ]; then
    # --- MACBOOK ROLE (Site 3) ---
    $APP --port 4446 -id 3 < /tmp/in_A3 > /tmp/out_A3 &
    $CTL -n C3 -nbsites 3 < /tmp/in_C3 > /tmp/out_C3 &

    cat /tmp/out_A3 > /tmp/in_C3 &

    # Network IN: Mac LISTENS for C2 coming from WSL
    nc -l $PORT_C2_TO_C3 > /tmp/in_C3 &
    
    # Network OUT: Mac LISTENS to send C3 to WSL
    cat /tmp/out_C3 | tee /tmp/in_A3 | nc -l $PORT_C3_TO_C1 &

    echo "Mac (Site 3) is running."
    echo "Waiting for WSL to connect on ports $PORT_C2_TO_C3 and $PORT_C3_TO_C1..."

elif [ "$ROLE" == "wsl" ]; then
    # --- WSL ROLE (Sites 1 & 2) ---
    $APP --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
    $CTL -n C1 -nbsites 3 < /tmp/in_C1 > /tmp/out_C1 &
    $APP --port 4445 -id 2 < /tmp/in_A2 > /tmp/out_A2 &
    $CTL -n C2 -nbsites 3 < /tmp/in_C2 > /tmp/out_C2 &

    # Site 1 Local Routing
    cat /tmp/out_A1 > /tmp/in_C1 &
    cat /tmp/out_C1 | tee /tmp/in_A1 > /tmp/in_C2 &
    
    # Site 2 Local Routing
    cat /tmp/out_A2 > /tmp/in_C2 &

    # Network OUT: WSL CONNECTS to Mac to send C2 to C3
    cat /tmp/out_C2 | tee /tmp/in_A2 | nc $MAC_IP $PORT_C2_TO_C3 &
    
    # Network IN: WSL CONNECTS to Mac to receive C3 into C1
    nc $MAC_IP $PORT_C3_TO_C1 > /tmp/in_C1 &

    echo "WSL (Sites 1 & 2) is running."
    echo "Connected to Mac at $MAC_IP."
fi

sleep 3600
#!/bin/bash

APP=$1
CTL=$2
ROLE=$3
MAC_IP=$4  # Only required when running the WSL role

PORT_C2_TO_C3=5001
PORT_C3_TO_C1=5002

# 1. Validate arguments and define FIFOs based on role
if [ "$ROLE" == "wsl" ]; then
    if [ -z "$MAC_IP" ]; then
        echo "Error: WSL role requires the Mac's IP address."
        echo "Usage: $0 <APP> <CTL> wsl <MAC_IP>"
        exit 1
    fi
    FIFOS=(/tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
           /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2)

elif [ "$ROLE" == "mac" ]; then
    FIFOS=(/tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3)

else
    echo "Usage: $0 <APP> <CTL> <wsl|mac> [MAC_IP]"
    exit 1
fi

# 2. Cleanup function
cleanup() {
    echo -e "\nCleaning up $ROLE..."
    kill $(jobs -p) 2>/dev/null
    rm -f "${FIFOS[@]}"
}
trap cleanup EXIT INT TERM

# 3. Create FIFOs
mkfifo "${FIFOS[@]}"


# 4. Execute topology based on Role
if [ "$ROLE" == "mac" ]; then
    # --- MACBOOK ROLE (Site 3) ---
    $APP --port 4446 -id 3 < /tmp/in_A3 > /tmp/out_A3 &
    $CTL -n C3 -nbsites 3 < /tmp/in_C3 > /tmp/out_C3 &

    cat /tmp/out_A3 > /tmp/in_C3 &

    # Network IN: Mac LISTENS for C2 coming from WSL
    nc -l $PORT_C2_TO_C3 > /tmp/in_C3 &
    
    # Network OUT: Mac LISTENS to send C3 to WSL
    cat /tmp/out_C3 | tee /tmp/in_A3 | nc -l $PORT_C3_TO_C1 &

    echo "Mac (Site 3) is running."
    echo "Waiting for WSL to connect on ports $PORT_C2_TO_C3 and $PORT_C3_TO_C1..."

elif [ "$ROLE" == "wsl" ]; then
    # --- WSL ROLE (Sites 1 & 2) ---
    $APP --port 4444 -id 1 < /tmp/in_A1 > /tmp/out_A1 &
    $CTL -n C1 -nbsites 3 < /tmp/in_C1 > /tmp/out_C1 &
    $APP --port 4445 -id 2 < /tmp/in_A2 > /tmp/out_A2 &
    $CTL -n C2 -nbsites 3 < /tmp/in_C2 > /tmp/out_C2 &

    # Site 1 Local Routing
    cat /tmp/out_A1 > /tmp/in_C1 &
    cat /tmp/out_C1 | tee /tmp/in_A1 > /tmp/in_C2 &
    
    # Site 2 Local Routing
    cat /tmp/out_A2 > /tmp/in_C2 &

    # Network OUT: WSL CONNECTS to Mac to send C2 to C3
    cat /tmp/out_C2 | tee /tmp/in_A2 | nc $MAC_IP $PORT_C2_TO_C3 &
    
    # Network IN: WSL CONNECTS to Mac to receive C3 into C1
    nc $MAC_IP $PORT_C3_TO_C1 > /tmp/in_C1 &

    echo "WSL (Sites 1 & 2) is running."
    echo "Connected to Mac at $MAC_IP."
fi

sleep 3600
