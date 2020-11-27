# CS5741--Project
Supermarket Checkout

I haven't merged this to main without checking, lets have a look at it tomorrow.

There is some issue with the last few shoppers of the day

Customer  169 is paying at Checkout  3...
Checkout 4;SCANNING;Prod.id;  6;SimScanTime;9.90;
Checkout 4;SCANNING;Prod.id; 28;SimScanTime;2.80;
Customer  175 is paying at Checkout  4...
Customer  169 is finished at Checkout  3.
Customer  175 is finished at Checkout  4.
Customer  190 arrived at Checkout  4 with  50 items
Checkout 4;SCANNING;Prod.id; 26;SimScanTime;4.80;
Customer  191 arrived at Checkout  5 with  50 items
Checkout 5;SCANNING;Prod.id; 23;SimScanTime;2.90;

The last 2 here arrive, start scanning but never finish?


More odd stuff now that we have a clock for the simulation
09:42:19:Customer  461 is paying at Checkout  6...
09:42:42:Checkout 4;SCANNING;Prod.id;107;SimScanTime;6.40;
09:42:56:Checkout 4;SCANNING;Prod.id;  3;SimScanTime;4.00;
09:43:06:Checkout 4;SCANNING;Prod.id; 50;SimScanTime;0.60;
09:43:08:Checkout 4;SCANNING;Prod.id; 32;SimScanTime;2.60;
09:43:15:Checkout 4;SCANNING;Prod.id;100;SimScanTime;3.60;
09:43:20:Customer  461 is finished at Checkout  6.
09:43:23:Checkout 4;SCANNING;Prod.id; 22;SimScanTime;5.30;
09:43:36:Checkout 4;SCANNING;Prod.id; 60;SimScanTime;5.60;
09:43:48:Checkout 4;SCANNING;Prod.id;109;SimScanTime;9.90;
09:44:10:Checkout 4;SCANNING;Prod.id; 16;SimScanTime;8.70;
09:44:30:Checkout 4;SCANNING;Prod.id; 20;SimScanTime;7.30;
09:44:47:Checkout 4;SCANNING;Prod.id; 45;SimScanTime;8.90;
09:45:07:Checkout 4;SCANNING;Prod.id; 65;SimScanTime;1.40;
09:45:10:Checkout 4;SCANNING;Prod.id; 19;SimScanTime;6.10;
09:45:24:Checkout 4;SCANNING;Prod.id; 69;SimScanTime;1.60;
09:45:28:Checkout 4;SCANNING;Prod.id; 90;SimScanTime;3.80;
09:45:38:Checkout 4;SCANNING;Prod.id; 26;SimScanTime;5.90;
09:45:51:Checkout 4;SCANNING;Prod.id; 51;SimScanTime;5.30;
09:46:03:Checkout 4;SCANNING;Prod.id; 52;SimScanTime;1.70;
09:46:07:Checkout 4;SCANNING;Prod.id; 56;SimScanTime;1.80;
09:46:12:Checkout 4;SCANNING;Prod.id; 76;SimScanTime;3.50;
09:46:20:Checkout 4;SCANNING;Prod.id; 88;SimScanTime;1.70;
09:46:24:Checkout 4;SCANNING;Prod.id; 96;SimScanTime;9.60;
09:46:46:Checkout 4;SCANNING;Prod.id;103;SimScanTime;10.00;
09:47:08:Customer  439 is paying at Checkout  4...
09:48:08:Customer  439 is finished at Checkout  4.
09:48:08:Customer  326 arrived at Checkout  4 with  51 items
09:48:08:Checkout 4;SCANNING;Prod.id;  7;SimScanTime;1.20;
09:48:08:Customer  466 arrived at Checkout  2 with 103 items
09:48:08:Checkout 2;SCANNING;Prod.id; 52;SimScanTime;4.60;
09:48:08:Customer  366 arrived at Checkout  8 with   9 items
09:48:08:Checkout 8;SCANNING;Prod.id;  1;SimScanTime;5.70;
09:48:08:Customer  402 arrived at Checkout  9 with 115 items
09:48:08:Checkout 9;SCANNING;Prod.id; 76;SimScanTime;3.10;
 
Look at 
09:47:08:Customer  439 is paying at Checkout  4...
09:48:08:Customer  439 is finished at Checkout  4.
Everything else seems blocked here, no other checkout does anything will this guy is paying.
hmmmmm...
