@echo off

javac -d ./bin -cp ./src;./lib/json.jar ./src/PointSalad.java


javac -d ./bin -cp ./src;./lib/org.junit4-4.3.1.jar;./lib/json.jar --release 17 ./src/PointSaladTest.java