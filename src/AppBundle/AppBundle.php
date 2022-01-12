<?php

namespace AppBundle;

use Symfony\Component\HttpKernel\Bundle\Bundle;

class AppBundle extends Bundle
{
    public const DATEFORMAT = [
        'month' => 'Y-m',
        'day'   => 'Y-m-d',
        'hour'  => 'Y-m-d H'
    ];

    public const VIEWFORMAT = [
        0 => [
            'month' => '%Y-%m-%d',
            'day'   => '%Y-%m-%d %H:00',
            'hour'  => '%Y-%m-%d %H:%M'
        ],
        1 => [
            'month' => '%Y-%m',
            'day'   => '%Y-%m-%d',
            'hour'  => '%Y-%m-%d %H'
        ]
    ];
}
