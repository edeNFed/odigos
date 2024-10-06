@@ -54,7 +54,6 @@ func updateInstrumentationInstanceStatus(status odigosv1.InstrumentationInstance
 }
 
 func InstrumentationInstanceName(ownerName string, pid int) string {
-
 	return fmt.Sprintf("%s-%d", ownerName, pid)
 }
 