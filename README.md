░█▀▀█ █▀▀ █▀▀ ░█─── █▀▀█ █▀▀█ █▀▀▄ ░█▀▄▀█ █▀▀█ █▀▀ █─█  
░█▀▀▄ █▀▀ █▀▀ ░█─── █──█ █▄▄█ █──█ ░█░█░█ █──█ █── █▀▄  
░█▄▄█ ▀▀▀ ▀▀▀ ░█▄▄█ ▀▀▀▀ ▀──▀ ▀▀▀─ ░█──░█ ▀▀▀▀ ▀▀▀ ▀─▀  

TOUCH IT
1) Build main package
2) Launch main executable file
3) Use :8000 to handle load request (URLs can be found in "mock" folder)
4) Use :7999 to send service requests    
    
Watch URLs in mock folder and try to send some requests to :8000    

FODLER STRUCTURE    
* executable file    
* main-config.yaml    
* service-config.yaml    
* some-cert.pub    
* some-cert.priv    
* mocks_folder   
    *   mock1.yaml   
    *   mock2.yaml   
    *   ...   

    
SERVICE REQUESTS    
* GET /shutdown    
  _stops :8000 handler_    
* GET /start    
  _launch :8000 handler_    
* GET /reboot    
  _shutdown + start_    
* POST /add-mock-file?path=optional_mock_folder&name=new.yaml && body = raw yaml file     
  _save new file in root_folder/optional_mock_folder/new.yaml_    
* POST /add-cert-file?path=optional_folder&name=new.pub && body = raw key file    
  _save new mock file in mock_folder/optional_folder/new.pub_    
* GET /remove-mock-file?uuid=08287033-2553-4556-a438-c2fac05e6553    
  _deletes first file in mock_folder that has identificator like 08287033-2553-4556-a438-c2fac05e6553_    
* GET /update-mainconfig-value?parameter=parameter_name&newvalue=new_param_value    
  _sets new_param_value to parameter_name_    
* GET /get-mainmock-config    
  _check main config_    

 FUNCTIONS 
 * $[uuid] - generates uuidV4    
 * $[timeNowFormatted(Monday, 02-Jan-06 15:04:05 MST)] - time.now() with custom format    
   _https://yourbasic.org/golang/format-parse-string-time-date-example/_    
 * $[randomString(0123456789;;10)] - random string of given symbols (digits for the case) and given lenght (10 here)
 * $[file(body.txt)] - copy the text from the file to mock yaml file, so you can keep huge bodies separated
    
How to gerexp here - https://github.com/google/re2/wiki/Syntax    
