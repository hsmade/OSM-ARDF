#!/usr/bin/env python3.7
# This script produces test messages

import math
from datetime import datetime


def calculate_bearing(a, b):
    lat1 = math.radians(a[1])
    lat2 = math.radians(b[1])

    diff_long = math.radians(b[0] - a[0])

    x = math.sin(diff_long) * math.cos(lat2)
    y = math.cos(lat1) * math.sin(lat2) - (math.sin(lat1)
                                           * math.cos(lat2) * math.cos(diff_long))

    initial_bearing = math.atan2(x, y)

    initial_bearing = math.degrees(initial_bearing)
    compass_bearing = (initial_bearing + 360) % 360

    return compass_bearing


def send_measurement(name, point, bearing):
    print('{{"timestamp":"{timestamp}", "station":"{station}", "longitude": {lon}, "latitude": {lat}, "bearing": {bearing}}}'.format(
        timestamp=datetime.utcnow().isoformat("T") + "Z",
        station=name,
        lon=point[0],
        lat=point[1],
        bearing=int(bearing))
    )


def main():
    car1 = [[float(item.split(',')[0]), float(item.split(',')[1])] for item in open("car1.coordinates").readlines()]
    car2 = [[float(item.split(',')[0]), float(item.split(',')[1])] for item in open("car2.coordinates").readlines()]
    balloon = [[float(item.split(',')[0]), float(item.split(',')[1])] for item in
               open("balloon.coordinates").readlines()]
    total = len(car1) + len(car2) + len(balloon)

    step_car1 = len(car1) / (total / 3.0)
    step_car2 = len(car2) / (total / 3.0)
    step_balloon = len(balloon) / (total / 3.0)

    index_car1 = index_car2 = index_balloon = 0

    while index_car1 < len(car1):
        bearing_car1 = calculate_bearing(car1[int(index_car1)], balloon[int(index_balloon) - 1])
        bearing_car2 = calculate_bearing(car2[int(index_car2)], balloon[int(index_balloon) - 1])
        send_measurement("car1", car1[int(index_car1)], bearing_car1)
        send_measurement("car2", car2[int(index_car2)], bearing_car2)
        send_measurement("balloon", balloon[int(index_balloon)], -1)

        index_car1 += step_car1
        index_car2 += step_car2
        index_balloon += step_balloon


if __name__ == "__main__":
    main()
