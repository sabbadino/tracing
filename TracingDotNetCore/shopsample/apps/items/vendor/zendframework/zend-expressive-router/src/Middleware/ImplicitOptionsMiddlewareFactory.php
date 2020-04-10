<?php
/**
 * @see       https://github.com/zendframework/zend-expressive-router for the canonical source repository
 * @copyright Copyright (c) 2018 Zend Technologies USA Inc. (https://www.zend.com)
 * @license   https://github.com/zendframework/zend-expressive-router/blob/master/LICENSE.md New BSD License
 */

declare(strict_types=1);

namespace Zend\Expressive\Router\Middleware;

use Psr\Container\ContainerInterface;
use Psr\Http\Message\ResponseInterface;
use Zend\Expressive\Router\Exception\MissingDependencyException;

/**
 * Create and return an ImplicitOptionsMiddleware instance.
 *
 * This factory depends on one other service:
 *
 * - Psr\Http\Message\ResponseInterface, which should resolve to a callable
 *   that will produce an empty Psr\Http\Message\ResponseInterface instance.
 */
class ImplicitOptionsMiddlewareFactory
{
    /**
     * @throws MissingDependencyException if the Psr\Http\Message\ResponseInterface
     *     service is missing.
     */
    public function __invoke(ContainerInterface $container) : ImplicitOptionsMiddleware
    {
        if (! $container->has(ResponseInterface::class)) {
            throw MissingDependencyException::dependencyForService(
                ResponseInterface::class,
                ImplicitOptionsMiddleware::class
            );
        }

        return new ImplicitOptionsMiddleware(
            $container->get(ResponseInterface::class)
        );
    }
}
