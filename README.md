# GOLANG
Metricbeats experience with Go and Kafka
I will share my experience building my Kafka module based on Metricbeats.
Based on the https://www.elastic.co/guide/en/beats/metricbeat/5.6/creating-beat-from-metricbeat.html

# Requirements 
- Go installed and in your PATH 
- Other requirements are as indicated on the above source page

During building or compiling of the Golang code, you might have to manually clone and copy required repositories to your ../beat_name/vendor/ directory.

During creation of the module, you will be prompted for the metricset name(check supported metricsets for the specific module). For the kafka case, we have two metricsets, start by adding one(say partition metricset) and run the command again to add another metricset but enter the same module name.
- This process will generate a template golang file for each metricset, which file you ll extend based on your needs.
