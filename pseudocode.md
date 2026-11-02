on initialised do
state := RELEASED;
End on

on enter do
 state := WANTED;
 "multicast 'req(T, p)'", where T := LAMPORT time of 'req' at p
 wait for N-1 replies
 state := HELD;
 End on

 on receive 'req(T_i, p_i)' do
 if (state == HELD || (state == WANTED && (T, P_me) < (T_i, p_i)))
 then queue req
 else reply to req
 End on

 On exit do
 state := RELEASED;
 reply to all in queue
 End on