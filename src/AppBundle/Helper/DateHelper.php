<?php

namespace AppBundle\Helper;

use AppBundle\AppBundle;
use DateTime;

class DateHelper
{
    /**
     * @param string $date
     * @param string $format
     * @return bool
     */
    public static function validateDate(string $date, string $format = 'Y-m-d H:i:s')
    {
        $dateTime = DateTime::createFromFormat($format, $date);
        return $dateTime && $dateTime->format($format) === $date;
    }

    /**
     * @param string $date
     * @return false|string
     */
    public static function analyzeDate(string $date)
    {
        if (self::validateDate($date, AppBundle::DATEFORMAT['month'])) {
            return 'month';
        }

        if (self::validateDate($date, AppBundle::DATEFORMAT['day'])) {
            return 'day';
        }

        if (self::validateDate($date, AppBundle::DATEFORMAT['hour'])) {
            return 'hour';
        }

        return false;
    }
}
