<?php

namespace AppBundle\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;
use Typomedia\Sysinfo\Exception\UnsupportedSystemException;
use Typomedia\Sysinfo\SysinfoFactory;

/**
 * @Route("/about")
 */
class AboutController extends AbstractController
{
    /**
     * @Route("/")
     * @throws UnsupportedSystemException
     */
    public function info(): Response
    {
        $sysinfo = SysinfoFactory::create();

        $system['os'] = $sysinfo->getOsRelease();
        $system['cpu'] = $sysinfo->getCpuModel();
        $system['ram'] = $sysinfo->getTotalMem();
        $system['php'] = $sysinfo->getPhpVersion();

        return $this->render('@App/sites/about.html.twig', ['system' => $system]);
    }
}
